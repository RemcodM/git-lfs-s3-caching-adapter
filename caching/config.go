package caching

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go/logging"
	"github.com/git-lfs/git-lfs/v3/config"
)

type cachingConfiguration struct {
	Bucket             *string  `json:"bucket,omitempty"`
	ConfigurationFiles []string `json:"configurationFiles,omitempty"`
	CredentialsFiles   []string `json:"credentialsFiles,omitempty"`
	Endpoint           *string  `json:"endpoint,omitempty"`
	Prefix             *string  `json:"prefix,omitempty"`
	Profile            *string  `json:"profile,omitempty"`
	Region             *string  `json:"region,omitempty"`
	Scope              *string  `json:"scope,omitempty"`
	UsePathStyle       *bool    `json:"usePathStyle,omitempty"`
}

func GetCachingConfiguration(cfg *config.Configuration) *cachingConfiguration {
	workingDir := cfg.LocalWorkingDir()

	cachingConfiguration := &cachingConfiguration{}
	_, err := os.Stat(workingDir + "/.lfscaching.json")
	if err == nil {
		file, err := os.Open(workingDir + "/.lfscaching.json")
		if err == nil {
			err = json.NewDecoder(file).Decode(cachingConfiguration)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error while decoding .lfsconfig.json. Will ignore its values\n")
			}
		} else {
			fmt.Fprintf(os.Stderr, "Error while opening .lfsconfig.json. Will ignore its values\n")
		}
	} else if !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error while checking existance of .lfsconfig.json. Will ignore its values\n")
	}

	if cachingConfiguration.Scope == nil {
		if value, ok := cfg.Git.Get("lfscache.scope"); ok {
			cachingConfiguration.Scope = &value
		}
	}

	scopes := []string{""}
	if cachingConfiguration.Scope != nil {
		scopes = append(scopes, fmt.Sprintf(".%s", *cachingConfiguration.Scope))
	}
	for _, scope := range scopes {
		if cachingConfiguration.Bucket == nil {
			if value, ok := cfg.Git.Get(fmt.Sprintf("lfscache%s.bucket", scope)); ok {
				cachingConfiguration.Bucket = &value
			} else {
				fmt.Fprintf(os.Stderr, "No bucket specified. Will not use caching\n")
				return nil
			}
		}
		if cachingConfiguration.ConfigurationFiles == nil {
			if values := cfg.Git.GetAll(fmt.Sprintf("lfscache%s.configFile", scope)); len(values) > 0 {
				cachingConfiguration.ConfigurationFiles = append(cachingConfiguration.ConfigurationFiles, values...)
			}
		}
		if cachingConfiguration.CredentialsFiles == nil {
			if values := cfg.Git.GetAll(fmt.Sprintf("lfscache%s.credentialsFile", scope)); len(values) > 0 {
				cachingConfiguration.CredentialsFiles = append(cachingConfiguration.CredentialsFiles, values...)
			}
		}
		if cachingConfiguration.Endpoint == nil {
			if value, ok := cfg.Git.Get(fmt.Sprintf("lfscache%s.endpoint", scope)); ok {
				cachingConfiguration.Endpoint = &value
			}
		}
		if cachingConfiguration.Prefix == nil {
			if value, ok := cfg.Git.Get(fmt.Sprintf("lfscache%s.prefix", scope)); ok {
				cachingConfiguration.Prefix = &value
			}
		}
		if cachingConfiguration.Profile == nil {
			if value, ok := cfg.Git.Get(fmt.Sprintf("lfscache%s.profile", scope)); ok {
				cachingConfiguration.Profile = &value
			}
		}
		if cachingConfiguration.Region == nil {
			if value, ok := cfg.Git.Get(fmt.Sprintf("lfscache%s.region", scope)); ok {
				cachingConfiguration.Region = &value
			}
		}
		if cachingConfiguration.UsePathStyle == nil {
			usePathStyle := cfg.Git.Bool(fmt.Sprintf("lfscache%s.usePathStyle", scope), false)
			cachingConfiguration.UsePathStyle = &usePathStyle
		}
	}

	return cachingConfiguration
}

func (c *cachingConfiguration) enabled() bool {
	return c.Bucket != nil
}

func (c *cachingConfiguration) newClient() (*s3.Client, error) {
	opts := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithLogger(logging.NewStandardLogger(os.Stderr)),
		awsconfig.WithClientLogMode(aws.LogRequest | aws.LogResponse),
	}
	if len(c.ConfigurationFiles) > 0 {
		opts = append(opts, awsconfig.WithSharedConfigFiles(c.ConfigurationFiles))
	}
	if len(c.CredentialsFiles) > 0 {
		opts = append(opts, awsconfig.WithSharedCredentialsFiles(c.CredentialsFiles))
	}
	if c.Profile != nil {
		opts = append(opts, awsconfig.WithSharedConfigProfile(*c.Profile))
	}

	config, err := awsconfig.LoadDefaultConfig(context.Background(), opts...)
	if err != nil {
		return nil, err
	}

	return s3.NewFromConfig(config, func(o *s3.Options) {
		if c.Endpoint != nil {
			o.BaseEndpoint = c.Endpoint
		}
		o.UsePathStyle = *c.UsePathStyle
	}), nil
}
