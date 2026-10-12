package longo

import (
	"context"
	"errors"
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Esquema cria a tabela de memórias, o índice e a política de isolamento por usuário.
const Esquema = `
CREATE EXTENSION IF NOT EXISTS vector;
CREATE TABLE IF NOT EXISTS memorias (
	id          bigserial   PRIMARY KEY,
	usuario     text        NOT NULL,
	tipo        text        NOT NULL,
	conteudo    text        NOT NULL,
	embedding   vector(256)  NOT NULL,
	importancia real        NOT NULL DEFAULT 0.5,
	criada_em   timestamptz NOT NULL DEFAULT now(),
	usada_em    timestamptz NOT NULL DEFAULT now(),
	usos        int         NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS memorias_usuario_idx ON memorias (usuario);
`

// Memoria é um item recuperado.
type Memoria struct {
	ID          int64
	Usuario     string
	Tipo        string
	Conteudo    string
	Importancia float32
	CriadaEm    time.Time
	UsadaEm     time.Time
	Usos        int
	Relevancia  float64 // similaridade de cosseno com a consulta, 0..1
	Pontuacao   float64 // relevância + recência + importância
}

// Loja é a memória de longo prazo.
type Loja struct {
	Pool     *pgxpool.Pool
	Emb      Embedder
	Dedup    float64       // similaridade a partir da qual um fato novo é considerado repetido
	MeiaVida time.Duration // a recência cai pela metade a cada MeiaVida
	Agora    func() time.Time

	// SemTrava desliga a trava por usuário do Guardar. Existe só para o
	// experimento que mostra o defeito; não use em produção.
	SemTrava bool
}

func (l *Loja) agora() time.Time {
	if l.Agora != nil {
		return l.Agora()
	}
	return time.Now()
}

// literal formata um vetor no texto que o pgvector aceita: [0.1,0.2,...].
func literal(v []float32) string {
	var b strings.Builder
	b.WriteByte('[')
	for i, x := range v {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(strconv.FormatFloat(float64(x), 'g', 8, 32))
	}
	b.WriteByte(']')
	return b.String()
}

// comUsuario roda fn em uma transação em que app.usuario vale usuario. A política
// de isolamento do banco (RLS) lê essa configuração.
func (l *Loja) comUsuario(ctx context.Context, usuario string, fn func(pgx.Tx) error) error {
	return pgx.BeginFunc(ctx, l.Pool, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `SELECT set_config('app.usuario', $1, true)`, usuario); err != nil {
			return err
		}
		return fn(tx)
	})
}

// Guardar grava uma memória. Se já existe uma muito parecida do mesmo usuário
// (similaridade >= Dedup), reforça a existente em vez de duplicar.
func (l *Loja) Guardar(ctx context.Context, usuario, tipo, conteudo string, importancia float32) (id int64, novo bool, err error) {
	v, err := l.Emb.Embed(ctx, conteudo)
	if err != nil {
		return 0, false, fmt.Errorf("embedding: %w", err)
	}
	lit := literal(v)
	err = l.comUsuario(ctx, usuario, func(tx pgx.Tx) error {
		if !l.SemTrava {
			// "procurar parecida e, se não houver, inserir" não é atômico: duas
			// transações podem não achar nada ao mesmo tempo e inserir as duas.
			// A trava consultiva serializa os Guardar do MESMO usuário (e só deles).
			if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1, 0))`, usuario); err != nil {
				return err
			}
		}
		var sim float64
		err := tx.QueryRow(ctx, `
			SELECT id, 1 - (embedding <=> $2::vector)
			FROM memorias WHERE usuario = $1 AND tipo = $3
			ORDER BY embedding <=> $2::vector LIMIT 1`, usuario, lit, tipo).Scan(&id, &sim)
		if err == nil && sim >= l.Dedup {
			_, err = tx.Exec(ctx, `
				UPDATE memorias
				SET usos = usos + 1, usada_em = $2, importancia = GREATEST(importancia, $3)
				WHERE id = $1`, id, l.agora(), importancia)
			return err
		}
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		novo = true
		return tx.QueryRow(ctx, `
			INSERT INTO memorias (usuario, tipo, conteudo, embedding, importancia, criada_em, usada_em)
			VALUES ($1, $2, $3, $4::vector, $5, $6, $6) RETURNING id`,
			usuario, tipo, conteudo, lit, importancia, l.agora()).Scan(&id)
	})
	return id, novo, err
}

// Pontuar combina relevância, recência e importância (todas entre 0 e 1).
func Pontuar(relevancia float64, idade, meiaVida time.Duration, importancia float64) float64 {
	recencia := math.Exp2(-float64(idade) / float64(meiaVida))
	return 0.6*relevancia + 0.25*recencia + 0.15*importancia
}

// Buscar devolve as k memórias mais úteis para a consulta. Pede ao banco os
// 4k vizinhos mais próximos (o índice faz isso rápido) e reordena em Go com a
// pontuação completa.
func (l *Loja) Buscar(ctx context.Context, usuario, consulta string, k int) ([]Memoria, error) {
	v, err := l.Emb.Embed(ctx, consulta)
	if err != nil {
		return nil, fmt.Errorf("embedding: %w", err)
	}
	var cand []Memoria
	err = l.comUsuario(ctx, usuario, func(tx pgx.Tx) error {
		linhas, err := tx.Query(ctx, `
			SELECT id, usuario, tipo, conteudo, importancia, criada_em, usada_em, usos,
			       1 - (embedding <=> $2::vector) AS relevancia
			FROM memorias WHERE usuario = $1
			ORDER BY embedding <=> $2::vector LIMIT $3`, usuario, literal(v), 4*k)
		if err != nil {
			return err
		}
		cand, err = pgx.CollectRows(linhas, func(r pgx.CollectableRow) (Memoria, error) {
			var m Memoria
			err := r.Scan(&m.ID, &m.Usuario, &m.Tipo, &m.Conteudo, &m.Importancia, &m.CriadaEm, &m.UsadaEm, &m.Usos, &m.Relevancia)
			return m, err
		})
		return err
	})
	if err != nil {
		return nil, err
	}
	agora := l.agora()
	for i := range cand {
		cand[i].Pontuacao = Pontuar(cand[i].Relevancia, agora.Sub(cand[i].UsadaEm), l.MeiaVida, float64(cand[i].Importancia))
	}
	slices.SortStableFunc(cand, func(a, b Memoria) int {
		switch {
		case a.Pontuacao > b.Pontuacao:
			return -1
		case a.Pontuacao < b.Pontuacao:
			return 1
		}
		return 0
	})
	if len(cand) > k {
		cand = cand[:k]
	}
	return cand, nil
}

// Esquecer apaga TODAS as memórias de um usuário e devolve quantas eram.
func (l *Loja) Esquecer(ctx context.Context, usuario string) (n int64, err error) {
	err = l.comUsuario(ctx, usuario, func(tx pgx.Tx) error {
		ct, err := tx.Exec(ctx, `DELETE FROM memorias WHERE usuario = $1`, usuario)
		n = ct.RowsAffected()
		return err
	})
	return n, err
}
