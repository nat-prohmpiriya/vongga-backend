package adapter

import (
	"context"

	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"

	"vongga_api/config"
)

func NewFirebaseProvider(config *config.Config) (*firebase.App, error) {
	opt := option.WithCredentialsFile(config.FirebaseCredentialsPath)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return nil, err
	}
	return app, nil
}
