package cfgloader

import (
	"context"
	"os"
	"reflect"
	"sync"
	"time"

	gsm "cloud.google.com/go/secretmanager/apiv1"
	gsmpb "cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/caarlos0/env/v11"
	"github.com/cockroachdb/errors"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"golang.org/x/sync/errgroup"
)

const (
	ConfigSecretPathTag = "envSecretPath"
)

var (
	globalValidator = validator.New()
)

type CallbackableItf interface {
	AfterLoad()
}

func Load[T any](ctx context.Context) (*T, error) {
	// load .env
	dotenvFiles := []string{}
	pathFromEnv := os.Getenv("DOTENV_PATH")
	if pathFromEnv != "" {
		dotenvFiles = append(dotenvFiles, pathFromEnv)
	}
	godotenv.Load(dotenvFiles...)

	// read config
	envOpts := env.Options{
		FuncMap: map[reflect.Type]env.ParserFunc{
			reflect.TypeOf(time.Location{}): parseLocation,
		},
	}
	cfg := new(T)
	if err := env.ParseWithOptions(cfg, envOpts); err != nil {
		return nil, err
	}

	if err := loadSecret(ctx, cfg); err != nil {
		return nil, err
	}

	if itf, ok := any(cfg).(CallbackableItf); ok {
		itf.AfterLoad()
	}

	// validate config
	if err := globalValidator.Struct(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// TODO:
// Loading secret is slow which will increase the startup time (about 120ms for 3 secrets on local).
// We should consider mount the secret to the container instead in the future.
// If we do that, the startup time can be reduced to < 5ms.
func loadSecret(ctx context.Context, cfg interface{}) error {
	// initialize secret manager client
	var (
		once   sync.Once
		client *gsm.Client
		err    error
	)
	getClient := func(ctx context.Context) (*gsm.Client, error) {
		once.Do(func() {
			client, err = gsm.NewClient(ctx)
			if err != nil {
				err = errors.Wrap(err, "failed to create secret manager client")
			}
		})
		return client, err
	}
	defer func() {
		if client != nil {
			client.Close()
		}
	}()

	// load secrets
	cfgV := reflect.ValueOf(cfg).Elem()
	fields := reflect.VisibleFields(cfgV.Type())
	wg, wgCtx := errgroup.WithContext(ctx)

	for _, field := range fields {
		secretPathEnv := field.Tag.Get(ConfigSecretPathTag)
		if secretPathEnv == "" {
			continue
		}

		secretPath := os.Getenv(secretPathEnv)
		if secretPath == "" {
			continue
		}

		wg.Go(func() error {
			data, err := loadOneSecret(wgCtx, secretPath, getClient)
			if err != nil {
				return err
			}

			fieldV := cfgV.FieldByName(field.Name)
			switch fieldV.Interface().(type) {
			case string:
				fieldV.SetString(string(data))
			case []byte:
				fieldV.SetBytes(data)
			default:
				// if it's necessary to support other types, use third party library to decode data
				return errors.New("unsupported secret type")
			}

			return nil
		})
	}

	// wait for all secrets to be loaded
	if err := wg.Wait(); err != nil {
		return errors.Wrap(err, "failed to wait for secret loaders")
	}

	return nil
}

func loadOneSecret(ctx context.Context, secretPath string, lazyClient func(context.Context) (*gsm.Client, error)) ([]byte, error) {
	client, err := lazyClient(ctx)
	if err != nil {
		return nil, err
	}

	result, err := client.AccessSecretVersion(ctx, &gsmpb.AccessSecretVersionRequest{
		Name: secretPath,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to access secret version")
	}
	return result.Payload.Data, nil
}

// Note: location is not yet supported by current version of env package, remove this when it's supported
func parseLocation(v string) (interface{}, error) {
	loc, err := time.LoadLocation(v)
	if err != nil {
		return nil, env.ParseValueError{
			Msg: "unable to parse location",
			Err: err,
		}
	}
	return *loc, nil
}
