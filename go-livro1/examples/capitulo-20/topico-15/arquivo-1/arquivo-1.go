package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Pedido struct {
	ID       bson.ObjectID `bson:"_id,omitempty"`
	Cliente  string        `bson:"cliente"`
	Itens    []Item        `bson:"itens"`
	Total    int64         `bson:"total"`
	Status   string        `bson:"status"`
	CriadoEm time.Time     `bson:"criado_em"`
}

type Item struct {
	SKU   string `bson:"sku"`
	Qtd   int    `bson:"qtd"`
	Preco int64  `bson:"preco"`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	client, err := mongo.Connect(options.Client().ApplyURI("mongodb://127.0.0.1:57017").SetMaxPoolSize(20))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal(err)
	}

	col := client.Database("livro").Collection("pedidos")
	col.Drop(ctx)

	// indice unico composto
	nome, err := col.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "cliente", Value: 1}, {Key: "criado_em", Value: -1}},
	})
	fmt.Println("indice:", nome, err)

	// Create
	p := Pedido{
		Cliente:  "ana",
		Itens:    []Item{{"TEC-1", 1, 25990}, {"MOU-2", 2, 8990}},
		Total:    25990 + 2*8990,
		Status:   "novo",
		CriadoEm: time.Now().UTC(),
	}
	r, err := col.InsertOne(ctx, p)
	if err != nil {
		log.Fatal(err)
	}
	id := r.InsertedID.(bson.ObjectID)
	fmt.Println("inserido:", id.Hex()[:8]+"...", len(id.Hex()))

	many, err := col.InsertMany(ctx, []Pedido{
		{Cliente: "bruno", Itens: []Item{{"MON-3", 1, 129900}}, Total: 129900, Status: "pago", CriadoEm: time.Now().UTC()},
		{Cliente: "ana", Itens: []Item{{"CAB-4", 3, 1990}}, Total: 5970, Status: "pago", CriadoEm: time.Now().UTC()},
	})
	fmt.Println("varios:", len(many.InsertedIDs), err)

	// Read
	var lido Pedido
	err = col.FindOne(ctx, bson.D{{Key: "_id", Value: id}}).Decode(&lido)
	fmt.Printf("lido: %s itens=%d total=%d\n", lido.Cliente, len(lido.Itens), lido.Total)

	err = col.FindOne(ctx, bson.D{{Key: "cliente", Value: "zeca"}}).Decode(&lido)
	fmt.Println("zeca:", errors.Is(err, mongo.ErrNoDocuments))

	cur, err := col.Find(ctx, bson.D{{Key: "status", Value: "pago"}},
		options.Find().SetSort(bson.D{{Key: "total", Value: -1}}).SetLimit(10))
	if err != nil {
		log.Fatal(err)
	}
	var pagos []Pedido
	if err := cur.All(ctx, &pagos); err != nil {
		log.Fatal(err)
	}
	for _, x := range pagos {
		fmt.Println("pago:", x.Cliente, x.Total)
	}

	// Update com operadores
	up, err := col.UpdateOne(ctx,
		bson.D{{Key: "_id", Value: id}},
		bson.D{
			{Key: "$set", Value: bson.D{{Key: "status", Value: "pago"}}},
			{Key: "$push", Value: bson.D{{Key: "itens", Value: Item{"BRI-5", 1, 500}}}},
			{Key: "$inc", Value: bson.D{{Key: "total", Value: 500}}},
		})
	fmt.Println("update: casou", up.MatchedCount, "mudou", up.ModifiedCount, err)

	// FindOneAndUpdate devolve o documento
	var depois Pedido
	err = col.FindOneAndUpdate(ctx,
		bson.D{{Key: "_id", Value: id}},
		bson.D{{Key: "$set", Value: bson.D{{Key: "status", Value: "enviado"}}}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&depois)
	fmt.Println("depois:", depois.Status, depois.Total, len(depois.Itens))

	// Aggregation: faturamento por cliente
	cur, err = col.Aggregate(ctx, mongo.Pipeline{
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$cliente"},
			{Key: "pedidos", Value: bson.D{{Key: "$sum", Value: 1}}},
			{Key: "faturado", Value: bson.D{{Key: "$sum", Value: "$total"}}},
		}}},
		{{Key: "$sort", Value: bson.D{{Key: "faturado", Value: -1}}}},
	})
	if err != nil {
		log.Fatal(err)
	}
	var grupos []bson.M
	if err := cur.All(ctx, &grupos); err != nil {
		log.Fatal(err)
	}
	for _, g := range grupos {
		fmt.Println("grupo:", g["_id"], g["pedidos"], g["faturado"])
	}

	// Delete
	d, err := col.DeleteOne(ctx, bson.D{{Key: "_id", Value: id}})
	fmt.Println("removidos:", d.DeletedCount, err)

	// transacao em instancia unica
	sess, _ := client.StartSession()
	defer sess.EndSession(ctx)
	_, err = sess.WithTransaction(ctx, func(ctx context.Context) (any, error) {
		return col.InsertOne(ctx, bson.D{{Key: "x", Value: 1}})
	})
	fmt.Println("transacao:", err)
}
