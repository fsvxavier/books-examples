package catalogo

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// documento e a forma do dado NO BANCO; Produto e a forma do dominio.
type documento struct {
	SKU   string   `bson:"_id"` // o sku e a propria chave primaria
	Nome  string   `bson:"nome"`
	Preco int64    `bson:"preco"`
	Tags  []string `bson:"tags"`
}

type Mongo struct{ col *mongo.Collection }

func NovoMongo(ctx context.Context, col *mongo.Collection) (*Mongo, error) {
	if err := col.Drop(ctx); err != nil {
		return nil, err
	}
	_, err := col.Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{Key: "tags", Value: 1}}})
	return &Mongo{col: col}, err
}

func (r *Mongo) Criar(ctx context.Context, p Produto) error {
	_, err := r.col.InsertOne(ctx, documento{p.SKU, p.Nome, p.Preco, p.Tags})
	if mongo.IsDuplicateKeyError(err) {
		return ErrDuplicado
	}
	return err
}

func (r *Mongo) Buscar(ctx context.Context, sku string) (Produto, error) {
	var d documento
	err := r.col.FindOne(ctx, bson.D{{Key: "_id", Value: sku}}).Decode(&d)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return Produto{}, ErrNaoEncontrado
	}
	return Produto{d.SKU, d.Nome, d.Preco, d.Tags}, err
}

func (r *Mongo) AtualizarPreco(ctx context.Context, sku string, preco int64) error {
	res, err := r.col.UpdateOne(ctx,
		bson.D{{Key: "_id", Value: sku}},
		bson.D{{Key: "$set", Value: bson.D{{Key: "preco", Value: preco}}}})
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 { // MatchedCount, nao ModifiedCount: mesmo preco nao e "nao encontrado"
		return ErrNaoEncontrado
	}
	return nil
}

func (r *Mongo) Remover(ctx context.Context, sku string) error {
	res, err := r.col.DeleteOne(ctx, bson.D{{Key: "_id", Value: sku}})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return ErrNaoEncontrado
	}
	return nil
}

func (r *Mongo) ListarPorTag(ctx context.Context, tag string) ([]Produto, error) {
	cur, err := r.col.Find(ctx, bson.D{{Key: "tags", Value: tag}},
		options.Find().SetSort(bson.D{{Key: "_id", Value: 1}}))
	if err != nil {
		return nil, err
	}
	var docs []documento
	if err := cur.All(ctx, &docs); err != nil {
		return nil, err
	}
	out := make([]Produto, len(docs))
	for i, d := range docs {
		out[i] = Produto{d.SKU, d.Nome, d.Preco, d.Tags}
	}
	return out, nil
}
