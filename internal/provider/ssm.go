package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

// SSMClient defines the subset of the SSM API we need.
type SSMClient interface {
	GetParameter(ctx context.Context, params *ssm.GetParameterInput, optFns ...func(*ssm.Options)) (*ssm.GetParameterOutput, error)
	GetParametersByPath(ctx context.Context, params *ssm.GetParametersByPathInput, optFns ...func(*ssm.Options)) (*ssm.GetParametersByPathOutput, error)
}

// SSMProvider retrieves secrets from AWS SSM Parameter Store.
type SSMProvider struct {
	client SSMClient
	region string
}

// NewSSMProvider creates a new SSMProvider. If region is empty the default
// AWS region resolution chain is used.
func NewSSMProvider(ctx context.Context, region string) (*SSMProvider, error) {
	var opts []func(*config.LoadOptions) error
	if region != "" {
		opts = append(opts, config.WithRegion(region))
	}
	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("ssm: load aws config: %w", err)
	}
	return &SSMProvider{
		client: ssm.NewFromConfig(cfg),
		region: cfg.Region,
	}, nil
}

// GetSecret retrieves a single parameter. ref is the SSM parameter path,
// e.g. "/myapp/db/password".
func (p *SSMProvider) GetSecret(ctx context.Context, ref string) (string, error) {
	out, err := p.client.GetParameter(ctx, &ssm.GetParameterInput{
		Name:           aws.String(ref),
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		return "", fmt.Errorf("ssm: get parameter %q: %w", ref, err)
	}
	if out.Parameter == nil || out.Parameter.Value == nil {
		return "", fmt.Errorf("ssm: parameter %q returned nil value", ref)
	}
	return aws.ToString(out.Parameter.Value), nil
}

// GetSecretsByPath retrieves all parameters under a path prefix and returns
// them as a map of basename -> value, suitable for bulk env injection.
func (p *SSMProvider) GetSecretsByPath(ctx context.Context, path string) (map[string]string, error) {
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}
	result := make(map[string]string)
	var nextToken *string
	for {
		out, err := p.client.GetParametersByPath(ctx, &ssm.GetParametersByPathInput{
			Path:           aws.String(path),
			WithDecryption: aws.Bool(true),
			Recursive:      aws.Bool(false),
			NextToken:      nextToken,
		})
		if err != nil {
			return nil, fmt.Errorf("ssm: get parameters by path %q: %w", path, err)
		}
		for _, p := range out.Parameters {
			name := aws.ToString(p.Name)
			key := name[strings.LastIndex(name, "/")+1:]
			result[key] = aws.ToString(p.Value)
		}
		if out.NextToken == nil {
			break
		}
		nextToken = out.NextToken
	}
	return result, nil
}
