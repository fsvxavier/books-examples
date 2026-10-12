// Package mensageria monta o Publisher e o Subscriber do Watermill. É o único
// lugar do programa que sabe qual tecnologia de mensageria está em uso: as
// fatias recebem as interfaces message.Publisher e message.Subscriber.
package mensageria

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/redis/go-redis/v9"
)

// NovoPublisherRedis publica em Streams do Redis (ou do Valkey, que fala o
// mesmo protocolo). Cria um cliente PRÓPRIO: o Close do publisher do Watermill
// fecha o cliente que recebeu, então compartilhá-lo com outro componente é uma
// armadilha (veja o capítulo).
func NovoPublisherRedis(endereco string, log *slog.Logger) (message.Publisher, error) {
	cli, err := NovoCliente(endereco)
	if err != nil {
		return nil, err
	}
	pub, err := redisstream.NewPublisher(redisstream.PublisherConfig{Client: cli}, watermill.NewSlogLogger(log))
	if err != nil {
		cli.Close()
		return nil, fmt.Errorf("criando o publisher do Redis: %w", err)
	}
	return pub, nil
}

// NovoSubscriberRedis assina com um grupo de consumidores: cada mensagem vai a
// UM consumidor do grupo, e um grupo diferente recebe a sua própria cópia.
// Uma mensagem que o consumidor não confirma (Nack, ou queda) é reentregue.
// Também cria um cliente próprio, que o Close do subscriber fecha.
func NovoSubscriberRedis(endereco, grupo string, log *slog.Logger) (message.Subscriber, error) {
	cli, err := NovoCliente(endereco)
	if err != nil {
		return nil, err
	}
	sub, err := redisstream.NewSubscriber(redisstream.SubscriberConfig{
		Client:          cli,
		ConsumerGroup:   grupo,
		Consumer:        grupo + "-1",
		OldestId:        "0", // um grupo novo começa do início do fluxo
		NackResendSleep: 200 * time.Millisecond,
		BlockTime:       200 * time.Millisecond,
		ClaimInterval:   time.Second,
		MaxIdleTime:     5 * time.Second,
	}, watermill.NewSlogLogger(log))
	if err != nil {
		cli.Close()
		return nil, fmt.Errorf("criando o subscriber do Redis: %w", err)
	}
	return sub, nil
}

// NovoCanal devolve um Pub/Sub em memória, que é publisher e subscriber ao
// mesmo tempo. Serve aos testes: sem rede, sem Redis.
func NovoCanal(log *slog.Logger) *gochannel.GoChannel {
	return gochannel.NewGoChannel(gochannel.Config{OutputChannelBuffer: 64}, watermill.NewSlogLogger(log))
}

// NovoCanalOrdenado é o Pub/Sub em memória em que Publish só retorna depois que
// o assinante confirma a mensagem. A entrega passa a ser sequencial, e os testes
// que dependem da ORDEM podem contar com ela. O NovoCanal comum entrega cada
// mensagem em uma goroutine própria e não garante a ordem.
func NovoCanalOrdenado(log *slog.Logger) *gochannel.GoChannel {
	return gochannel.NewGoChannel(gochannel.Config{OutputChannelBuffer: 64, BlockPublishUntilSubscriberAck: true}, watermill.NewSlogLogger(log))
}

// descartar é um Publisher que joga fora o que recebe.
type descartar struct{}

func (descartar) Publish(string, ...*message.Message) error { return nil }
func (descartar) Close() error                              { return nil }

// Descartar devolve um Publisher que não publica nada, para os programas que
// não têm broker e não querem ter (a linha de comando, por exemplo).
func Descartar() message.Publisher { return descartar{} }

// NovoCliente abre um cliente do Redis (ou Valkey). O endereço pode ser
// "host:porta" ou uma URL (redis://host:porta/3, com o número do banco).
func NovoCliente(endereco string) (*redis.Client, error) {
	if strings.HasPrefix(endereco, "redis://") || strings.HasPrefix(endereco, "rediss://") {
		opcoes, err := redis.ParseURL(endereco)
		if err != nil {
			return nil, fmt.Errorf("endereço do Redis inválido: %w", err)
		}
		return redis.NewClient(opcoes), nil
	}
	return redis.NewClient(&redis.Options{Addr: endereco}), nil
}

// Resumo devolve só a primeira linha da mensagem de um erro. Os erros das
// bibliotecas de mensageria costumam trazer a pilha de chamadas dentro do
// próprio texto, e uma queda do broker, com milhares de tentativas, enche o log
// de pilhas idênticas. O resumo diz o que aconteceu; a pilha, se for preciso,
// fica no erro original.
func Resumo(err error) string {
	linha, _, _ := strings.Cut(err.Error(), "\n")
	return linha
}
