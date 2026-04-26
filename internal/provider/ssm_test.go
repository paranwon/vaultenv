package provider

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
)

// fakeSSMClient implements SSMClient for testing.
type fakeSSMClient struct {
	params map[string]string
}

func (f *fakeSSMClient) GetParameter(_ context.Context, in *ssm.GetParameterInput, _ ...func(*ssm.Options)) (*ssm.GetParameterOutput, error) {
	v, ok := f.params[aws.ToString(in.Name)]
	if !ok {
		return nil, &types.ParameterNotFound{}
	}
	return &ssm.GetParameterOutput{
		Parameter: &types.Parameter{Name: in.Name, Value: aws.String(v)},
	}, nil
}

func (f *fakeSSMClient) GetParametersByPath(_ context.Context, in *ssm.GetParametersByPathInput, _ ...func(*ssm.Options)) (*ssm.GetParametersByPathOutput, error) {
	prefix := aws.ToString(in.Path)
	var params []types.Parameter
	for k, v := range f.params {
		if len(k) > len(prefix) && k[:len(prefix)] == prefix {
			params = append(params, types.Parameter{
				Name:  aws.String(k),
				Value: aws.String(v),
			})
		}
	}
	return &ssm.GetParametersByPathOutput{Parameters: params}, nil
}

func TestSSMProvider_GetSecret(t *testing.T) {
	p := &SSMProvider{client: &fakeSSMClient{
		params: map[string]string{"/app/db/password": "s3cr3t"},
	}}
	val, err := p.GetSecret(context.Background(), "/app/db/password")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "s3cr3t" {
		t.Errorf("expected %q, got %q", "s3cr3t", val)
	}
}

func TestSSMProvider_GetSecret_NotFound(t *testing.T) {
	p := &SSMProvider{client: &fakeSSMClient{params: map[string]string{}}}
	_, err := p.GetSecret(context.Background(), "/missing")
	if err == nil {
		t.Fatal("expected error for missing parameter")
	}
}

func TestSSMProvider_GetSecretsByPath(t *testing.T) {
	p := &SSMProvider{client: &fakeSSMClient{
		params: map[string]string{
			"/app/prod/DB_HOST": "localhost",
			"/app/prod/DB_PORT": "5432",
			"/app/other/KEY":   "ignored",
		},
	}}
	secrets, err := p.GetSecretsByPath(context.Background(), "/app/prod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(secrets) != 2 {
		t.Fatalf("expected 2 secrets, got %d", len(secrets))
	}
	if secrets["DB_HOST"] != "localhost" {
		t.Errorf("DB_HOST: expected %q, got %q", "localhost", secrets["DB_HOST"])
	}
	if secrets["DB_PORT"] != "5432" {
		t.Errorf("DB_PORT: expected %q, got %q", "5432", secrets["DB_PORT"])
	}
}
