// Package expiracao é a fatia "expirar um link". Um link expira pela passagem do
// tempo, e ninguém o aciona: não há requisição quando o prazo termina. A fatia
// junta as duas ferramentas do capítulo: o Watermill diz que um link foi criado
// (e com que prazo), e o Asynq cuida do instante em que o prazo termina.
package expiracao

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/hibiken/asynq"

	"exemplos/livro4/capitulo-7/topico-2/encurtador/internal/infra/mensageria"
	"exemplos/livro4/capitulo-7/topico-2/encurtador/internal/mensagens"
)

// TipoExpirar é o tipo da tarefa do Asynq.
const TipoExpirar = "link:expirar"

type payloadExpirar struct {
	Codigo string `json:"codigo"`
}

// Servico é a fatia: agenda a expiração de cada link com prazo e, no instante
// certo, publica a mensagem link_expirado.
type Servico struct {
	tarefas *asynq.Client
	pub     message.Publisher
	log     *slog.Logger
}

// Novo monta a fatia.
func Novo(tarefas *asynq.Client, pub message.Publisher, log *slog.Logger) *Servico {
	return &Servico{tarefas: tarefas, pub: pub, log: log}
}

// Agendar é o consumidor da mensagem link_criado: se o link tem prazo, agenda
// uma tarefa para o fim dele. É idempotente: o TaskID é derivado do código, então
// uma mensagem repetida (a entrega do broker é "pelo menos uma vez") não cria uma
// segunda tarefa.
//
// O instante agendado é o fim do prazo MAIS UM SEGUNDO: o Asynq guarda o instante
// em segundos inteiros, truncando a fração, e sem a soma a tarefa poderia rodar
// até um segundo ANTES do prazo, anunciando como passado um fato que ainda não
// aconteceu.
func (s *Servico) Agendar(ctx context.Context, msg *message.Message) error {
	env, err := mensagens.DoBytes(msg.Payload)
	if err != nil {
		return err
	}
	criado, err := mensagens.Decodificar[mensagens.LinkCriadoV1](env)
	if err != nil {
		return err
	}
	if criado.Expira == nil {
		return nil // um link sem prazo não expira
	}
	payload, err := json.Marshal(payloadExpirar{Codigo: criado.Codigo})
	if err != nil {
		return err
	}
	_, err = s.tarefas.EnqueueContext(ctx, asynq.NewTask(TipoExpirar, payload),
		asynq.ProcessAt(criado.Expira.Add(time.Second)),
		asynq.TaskID("expirar:"+criado.Codigo),
		asynq.MaxRetry(5),
	)
	if errors.Is(err, asynq.ErrTaskIDConflict) {
		return nil // já agendada: a mensagem era uma repetição
	}
	if err != nil {
		return fmt.Errorf("agendando a expiração de %s: %w", criado.Codigo, err)
	}
	return nil
}

// Expirar é o manipulador da tarefa: chegou o instante, publica link_expirado.
// O identificador da mensagem é derivado do código, e não sorteado: o Asynq
// garante "pelo menos uma execução", então Expirar pode rodar duas vezes, e as
// duas execuções produzem a MESMA mensagem, que os consumidores já sabem descartar.
func (s *Servico) Expirar(_ context.Context, t *asynq.Task) error {
	var p payloadExpirar
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		// Um dado corrompido não melhora com novas tentativas.
		return fmt.Errorf("payload ilegível: %v: %w", err, asynq.SkipRetry)
	}
	env, err := mensagens.Novo("expirado:"+p.Codigo, "link_expirado", time.Now(), mensagens.LinkExpiradoV1{Codigo: p.Codigo})
	if err != nil {
		return err
	}
	bytes, err := env.Bytes()
	if err != nil {
		return err
	}
	if err := s.pub.Publish(mensagens.TopicoLinkExpirado, message.NewMessage(env.ID, bytes)); err != nil {
		s.log.Warn("expiração não publicada (o Asynq tenta de novo)", "codigo", p.Codigo, "erro", mensageria.Resumo(err))
		return err
	}
	return nil
}

// RegistrarConsumidor inscreve Agendar no roteador do Watermill, no tópico dos
// links criados.
func (s *Servico) RegistrarConsumidor(r *message.Router, sub message.Subscriber) {
	r.AddConsumerHandler("expiracao", mensagens.TopicoLinkCriado, sub, func(msg *message.Message) error {
		return s.Agendar(msg.Context(), msg)
	})
}

// RegistrarTarefa inscreve Expirar no ServeMux do Asynq.
func (s *Servico) RegistrarTarefa(mux *asynq.ServeMux) {
	mux.HandleFunc(TipoExpirar, s.Expirar)
}
