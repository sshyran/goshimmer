package mqtt

import (
	"encoding/json"

	"github.com/iotaledger/goshimmer/packages/tangle"
	"github.com/iotaledger/hive.go/identity"
)

func publishMessage(ev *tangle.CachedMessageEvent) {
	defer ev.MessageMetadata.Release()

	ev.Message.Consume(func(msg *tangle.Message) {

		if mqttBroker.HasSubscribers(topicMessagesSolid) {

			messageResponse := message{
				MessageID: msg.ID().String(),
				IssuerID:  identity.New(msg.IssuerPublicKey()).ID().String(),
				Timestamp: msg.IssuingTime().UnixNano(),
				Payload:   msg.Payload().Bytes(),
				Nonce:     msg.Nonce(),
				Signature: msg.Signature().String(),
			}

			msg.ForEachStrongParent(func(parent tangle.MessageID) {
				messageResponse.StrongParents = append(messageResponse.StrongParents, parent.String())
			})

			msg.ForEachWeakParent(func(parent tangle.MessageID) {
				messageResponse.WeakParents = append(messageResponse.WeakParents, parent.String())
			})

			// Serialize here instead of using publishOnTopic to avoid double JSON marshalling
			jsonPayload, err := json.Marshal(messageResponse)
			if err != nil {
				log.Warn(err.Error())
				return
			}

			mqttBroker.Send(topicMessagesSolid, jsonPayload)

		}
	})

}
