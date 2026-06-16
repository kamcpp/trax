package mq

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/xshyft/trax/pkg/common"
	mqcommon "github.com/xshyft/trax/pkg/mq/common"
	mqtrax "github.com/xshyft/trax/pkg/mq/trax"
)

func initQueues(ctx context.Context) error {
	if mqcommon.RabbitMQChannelPool == nil {
		return fmt.Errorf("rabbitmq channel pool not initialized")
	}

	ch, err := mqcommon.RabbitMQChannelPool.GetChannel()
	if err != nil {
		return fmt.Errorf("failed to get channel from pool: %w", err)
	}
	defer mqcommon.RabbitMQChannelPool.ReturnChannel(ch)

	if err := ch.Qos(1000, 0, false); err != nil {
		return err
	}
	return mqtrax.InitTraxIncomingSagasSystem(ctx)
}

func initConn(_ctx context.Context) error {
	if len(mqcommon.RabbitMQURL) == 0 {
		mqcommon.RabbitMQURL = os.Getenv("RABBITMQ_CONN_STRING")
		if len(mqcommon.RabbitMQURL) == 0 {
			panic("RABBITMQ_CONN_STRING is not set")
		}
	}
	config := amqp.Config{
		Heartbeat: 60 * time.Second,
		Dial: func(network, addr string) (net.Conn, error) {
			return amqp.DefaultDial(time.Second*300)(network, addr)
		},
	}
	var err error
	mqcommon.RabbitMQConnection, err = amqp.DialConfig(mqcommon.RabbitMQURL, config)
	if err != nil {
		return err
	}

	if mqcommon.RabbitMQChannelPool != nil && !mqcommon.RabbitMQChannelPool.IsClosed() {
		mqcommon.RabbitMQChannelPool.UpdateConnection(mqcommon.RabbitMQConnection)
	} else {
		if mqcommon.RabbitMQChannelPool != nil {
			mqcommon.RabbitMQChannelPool.Close()
		}

		maxChannels := 500
		if maxChannelsStr := os.Getenv("RABBITMQ_MAX_CHANNELS"); maxChannelsStr != "" {
			fmt.Sscanf(maxChannelsStr, "%d", &maxChannels)
		}

		mqcommon.RabbitMQChannelPool, err = mqcommon.NewChannelPool(mqcommon.RabbitMQConnection, maxChannels)
		if err != nil {
			return fmt.Errorf("failed to create channel pool: %w", err)
		}
	}

	return nil
}

func Init(ctx context.Context) {
	for {
		if err := initConn(ctx); err != nil {
			common.L.Warn(err.Error(), common.F(ctx)...)
			time.Sleep(10 * time.Second)
			continue
		}
		break
	}
	for {
		if err := initQueues(ctx); err != nil {
			common.L.Warn(err.Error(), common.F(ctx)...)
			time.Sleep(10 * time.Second)
			continue
		}
		break
	}
	common.L.Debug("rabbitmq connect success", common.F(ctx)...)
	go func() {
		for {
			closeCh := make(chan *amqp.Error)
			mqcommon.RabbitMQConnection.NotifyClose(closeCh)
			reason, ok := <-closeCh
			if !ok {
				common.L.Error("rabbitmq connection closed", common.F(ctx)...)
				break
			}
			common.L.Warn(fmt.Sprintf("rabbitmq connection closed unexpectedly, reason: %v", reason), common.F(ctx)...)

			attemptNum := 0
			for {
				attemptNum++
				time.Sleep(10 * time.Second)
				if err := initConn(ctx); err == nil {
					if err2 := initQueues(ctx); err2 == nil {
						mqcommon.NotifyReconnect()
						break
					}
				}
			}
		}
	}()
}
