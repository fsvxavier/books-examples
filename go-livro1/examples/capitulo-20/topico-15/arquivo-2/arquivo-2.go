package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Usuario struct {
	ID      bson.ObjectID  `bson:"_id,omitempty"`
	Email   string         `bson:"email"`
	Nome    string         `bson:"nome"`
	Perfil  map[string]any `bson:"perfil,omitempty"`
	Visitas int            `bson:"visitas"`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	client, err := mongo.Connect(options.Client().ApplyURI("mongodb://127.0.0.1:57017"))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)
	col := client.Database("livro").Collection("usuarios")
	col.Drop(ctx)

	col.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	})

	_, err = col.InsertOne(ctx, Usuario{Email: "ana@exemplo.com", Nome: "Ana"})
	fmt.Println("primeiro:", err)
	_, err = col.InsertOne(ctx, Usuario{Email: "ana@exemplo.com", Nome: "Outra Ana"})
	fmt.Println("duplicado:", mongo.IsDuplicateKeyError(err))

	// upsert: cria se nao existe, incrementa se existe
	for range 3 {
		r, err := col.UpdateOne(ctx,
			bson.D{{Key: "email", Value: "bia@exemplo.com"}},
			bson.D{
				{Key: "$setOnInsert", Value: bson.D{{Key: "nome", Value: "Bia"}}},
				{Key: "$inc", Value: bson.D{{Key: "visitas", Value: 1}}},
			},
			options.UpdateOne().SetUpsert(true))
		fmt.Println("upsert: casou", r.MatchedCount, "inseriu", r.UpsertedCount, err)
	}

	// documentos com formatos diferentes na mesma colecao
	col.InsertOne(ctx, Usuario{Email: "caio@exemplo.com", Nome: "Caio",
		Perfil: map[string]any{"cidade": "Recife", "interesses": []string{"go", "ciclismo"}}})

	// consulta em campo aninhado e projecao
	cur, _ := col.Find(ctx, bson.D{{Key: "perfil.cidade", Value: "Recife"}},
		options.Find().SetProjection(bson.D{{Key: "email", Value: 1}, {Key: "_id", Value: 0}}))
	var emails []bson.M
	cur.All(ctx, &emails)
	fmt.Println("recife:", emails)

	// BulkWrite
	res, err := col.BulkWrite(ctx, []mongo.WriteModel{
		mongo.NewUpdateOneModel().SetFilter(bson.D{{Key: "email", Value: "ana@exemplo.com"}}).
			SetUpdate(bson.D{{Key: "$set", Value: bson.D{{Key: "nome", Value: "Ana Maria"}}}}),
		mongo.NewDeleteOneModel().SetFilter(bson.D{{Key: "email", Value: "bia@exemplo.com"}}),
	})
	fmt.Println("bulk: mudou", res.ModifiedCount, "apagou", res.DeletedCount, err)

	n, _ := col.CountDocuments(ctx, bson.D{})
	fmt.Println("total:", n)

	// explain mostra se o indice foi usado
	var plano bson.Raw
	err = client.Database("livro").RunCommand(ctx, bson.D{
		{Key: "explain", Value: bson.D{
			{Key: "find", Value: "usuarios"},
			{Key: "filter", Value: bson.D{{Key: "email", Value: "ana@exemplo.com"}}},
		}},
		{Key: "verbosity", Value: "queryPlanner"},
	}).Decode(&plano)
	fmt.Println("plano:", plano.Lookup("queryPlanner", "winningPlan", "stage").StringValue(), plano.Lookup("queryPlanner", "winningPlan", "indexName").StringValue(), err)
}
