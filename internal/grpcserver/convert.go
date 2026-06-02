package grpcserver

import (
	"github.com/iliaonishchenko/gophkeeper/internal/model"
	pb "github.com/iliaonishchenko/gophkeeper/internal/proto"
)

func modelToProto(it *model.Item) *pb.Item {
	return &pb.Item{
		Id:         it.ID,
		Type:       pb.ItemType(it.Type),
		Name:       it.Name,
		Ciphertext: it.Ciphertext,
		Metadata:   it.Metadata,
		Version:    it.Version,
		UpdatedAt:  it.UpdatedAt,
		Deleted:    it.Deleted,
	}
}

func protoToModel(it *pb.Item) *model.Item {
	if it == nil {
		return &model.Item{}
	}
	return &model.Item{
		ID:         it.GetId(),
		Type:       model.ItemType(it.GetType()),
		Name:       it.GetName(),
		Ciphertext: it.GetCiphertext(),
		Metadata:   it.GetMetadata(),
		Version:    it.GetVersion(),
		UpdatedAt:  it.GetUpdatedAt(),
		Deleted:    it.GetDeleted(),
	}
}
