package gql

import (
	"context"

	"github.com/jackytck/alti-cli/config"
	"github.com/jackytck/alti-cli/errors"
	"github.com/machinebox/graphql"
)

var bucketType = map[string]map[string]string{
	"image": {
		"s3":    "BucketS3",
		"oss":   "BucketOSS",
		"minio": "BucketMinio",
	},
	"model": {
		"s3":    "BucketS3Model",
		"minio": "BucketMinioModel",
	},
}

// BucketList returns a list of available buckets supported by the api server.
// 'kind' is 'image' or 'model'.
// 'cloud' is 's3', 'oss' or 'minio'.
func BucketList(kind, cloud string) ([]string, error) {
	var ret []string

	b, ok := bucketType[kind]
	if !ok {
		return ret, errors.ErrBucketInvalid
	}
	t, ok := b[cloud]
	if !ok {
		return ret, errors.ErrBucketInvalid
	}

	config := config.Load()
	active := config.GetActive()
	client := graphql.NewClient(active.Endpoint + "/graphql")

	req := graphql.NewRequest(`
		query ($type: String!) {
			__type(name: $type) {
				enumValues {
					name
				}
			}
		}
	`)

	req.Header.Set("key", active.Key)
	req.Var("type", t)

	ctx := context.Background()
	var res bucketRes
	if err := client.Run(ctx, req, &res); err != nil {
		return ret, err
	}

	for _, b := range res.Type.EnumValues {
		ret = append(ret, b.Name)
	}

	return ret, nil
}

type bucketRes struct {
	Type enumBukType `json:"__type"`
}

type enumBukType struct {
	EnumValues []struct {
		Name string
	}
}
