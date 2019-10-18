package storage

import (
	"fmt"

	"github.com/nicolagi/dino/message"
	log "github.com/sirupsen/logrus"
)

func ApplyMessage(store VersionedStore, in message.Message) (out message.Message) {
	inTag := in.Tag()
	switch in.Kind() {
	case message.KindGet:
		version, value, err := store.Get([]byte(in.Key()))
		if err != nil {
			return message.NewErrorMessage(inTag, err.Error())
		}
		return message.NewPutMessage(inTag, in.Key(), string(value), version)
	case message.KindPut:
		err := store.Put(in.Version(), []byte(in.Key()), []byte(in.Value()))
		if err != nil {
			return message.NewErrorMessage(inTag, err.Error())
		}
		log.WithFields(log.Fields{
			"key":     fmt.Sprintf("%.10x", in.Key()),
			"version": in.Version(),
		}).Debug("Applied put message")
		return in
	case message.KindError:
		return message.NewErrorMessage(inTag, "error messages cannot be applied")
	default:
		return message.NewErrorMessage(inTag, "unknown message kind")
	}
}
