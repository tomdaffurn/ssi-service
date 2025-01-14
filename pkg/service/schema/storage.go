package schema

import (
	"context"
	"fmt"

	"github.com/goccy/go-json"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/TBD54566975/ssi-sdk/credential/schema"

	"github.com/tbd54566975/ssi-service/pkg/storage"

	"github.com/tbd54566975/ssi-service/internal/keyaccess"
)

const (
	namespace = "schema"
)

type StoredSchema struct {
	ID        string              `json:"id"`
	Schema    schema.VCJSONSchema `json:"schema"`
	SchemaJWT *keyaccess.JWT      `json:"token,omitempty"`
}

type Storage struct {
	db storage.ServiceStorage
}

func NewSchemaStorage(db storage.ServiceStorage) (*Storage, error) {
	if db == nil {
		return nil, errors.New("bolt db reference is nil")
	}
	return &Storage{db: db}, nil
}

func (ss *Storage) StoreSchema(ctx context.Context, schema StoredSchema) error {
	id := schema.ID
	if id == "" {
		err := errors.New("could not store schema without an ID")
		logrus.WithError(err).Error()
		return err
	}
	schemaBytes, err := json.Marshal(schema)
	if err != nil {
		errMsg := fmt.Sprintf("could not store schema: %s", id)
		logrus.WithError(err).Error(errMsg)
		return errors.Wrapf(err, errMsg)
	}
	return ss.db.Write(ctx, namespace, id, schemaBytes)
}

func (ss *Storage) GetSchema(ctx context.Context, id string) (*StoredSchema, error) {
	schemaBytes, err := ss.db.Read(ctx, namespace, id)
	if err != nil {
		errMsg := fmt.Sprintf("could not get schema: %s", id)
		logrus.WithError(err).Error(errMsg)
		return nil, errors.Wrapf(err, errMsg)
	}
	if len(schemaBytes) == 0 {
		err := fmt.Errorf("schema not found with id: %s", id)
		logrus.WithError(err).Error("could not get schema from storage")
		return nil, err
	}
	var stored StoredSchema
	if err := json.Unmarshal(schemaBytes, &stored); err != nil {
		errMsg := fmt.Sprintf("could not unmarshal stored schema: %s", id)
		logrus.WithError(err).Error(errMsg)
		return nil, errors.Wrapf(err, errMsg)
	}
	return &stored, nil
}

// ListSchemas attempts to get all stored schemas. It will return those it can even if it has trouble with some.
func (ss *Storage) ListSchemas(ctx context.Context) ([]StoredSchema, error) {
	gotSchemas, err := ss.db.ReadAll(ctx, namespace)
	if err != nil {
		errMsg := "could not get all schemas"
		logrus.WithError(err).Error(errMsg)
		return nil, errors.Wrap(err, errMsg)
	}
	if len(gotSchemas) == 0 {
		logrus.Info("no schemas to get")
		return nil, nil
	}
	var stored []StoredSchema
	for _, schemaBytes := range gotSchemas {
		var nextSchema StoredSchema
		if err := json.Unmarshal(schemaBytes, &nextSchema); err == nil {
			stored = append(stored, nextSchema)
		}
	}
	return stored, nil
}

func (ss *Storage) DeleteSchema(ctx context.Context, id string) error {
	if err := ss.db.Delete(ctx, namespace, id); err != nil {
		errMsg := fmt.Sprintf("could not delete schema: %s", id)
		logrus.WithError(err).Error(errMsg)
		return errors.Wrapf(err, errMsg)
	}
	return nil
}
