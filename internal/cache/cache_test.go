package cache_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/internal/cache"
	"github.com/thomaspoignant/go-feature-flag/testutils/testconvert"
	"testing"

	"github.com/thomaspoignant/go-feature-flag/internal/model"
	"github.com/thomaspoignant/go-feature-flag/internal/notifier"
)

func Test_FlagCacheNotInit(t *testing.T) {
	fCache := cache.New(nil)
	fCache.Close()
	_, err := fCache.GetFlag("test-flag")
	assert.Error(t, err, "We should have an error if the cache is not init")
}

func Test_GetFlagNotExist(t *testing.T) {
	fCache := cache.New(nil)
	_, err := fCache.GetFlag("not-exists-flag")
	assert.Error(t, err, "We should have an error if the flag does not exists")
}

func Test_FlagCache(t *testing.T) {
	yamlFile := []byte(`test-flag:
  rule: key eq "random-key"
  percentage: 100
  true: true
  false: false
  default: false
  trackEvents: false
`)

	jsonFile := []byte(`{
  "test-flag": {
    "rule": "key eq \"random-key\"",
    "percentage": 100,
    "true": true,
    "false": false,
    "default": false
  }
}
`)

	tomlFile := []byte(`[test-flag]
rule = "key eq \"random-key\""
percentage = 100.0
true = true
false = false
default = false
disable = false`)

	type args struct {
		loadedFlags []byte
	}
	tests := []struct {
		name       string
		args       args
		expected   map[string]model.FlagData
		wantErr    bool
		flagFormat string
	}{
		{
			name:       "Yaml valid",
			flagFormat: "yaml",
			args: args{
				loadedFlags: yamlFile,
			},
			expected: map[string]model.FlagData{
				"test-flag": {
					Disable:     nil,
					Rule:        testconvert.String("key eq \"random-key\""),
					Percentage:  testconvert.Float64(100),
					True:        testconvert.Interface(true),
					False:       testconvert.Interface(false),
					Default:     testconvert.Interface(false),
					TrackEvents: testconvert.Bool(false),
				},
			},
			wantErr: false,
		},
		{
			name:       "Yaml invalid file",
			flagFormat: "yaml",
			args: args{
				loadedFlags: []byte(`test-flag:
  rule: key eq "random-key"
  percentage: "toot"
  true: true
  false: false
  default: false
`),
			},
			wantErr: true,
		},
		{
			name: "JSON valid",
			args: args{
				loadedFlags: jsonFile,
			},
			flagFormat: "json",
			expected: map[string]model.FlagData{
				"test-flag": {
					Rule:       testconvert.String("key eq \"random-key\""),
					Percentage: testconvert.Float64(100),
					True:       testconvert.Interface(true),
					False:      testconvert.Interface(false),
					Default:    testconvert.Interface(false),
				},
			},
			wantErr: false,
		},
		{
			name:       "JSON invalid file",
			flagFormat: "json",
			args: args{
				loadedFlags: []byte(`{
  "test-flag": {
    "rule": "key eq \"random-key\"",
    "percentage": 100,
    "true": true,
    "false": false,
    "default": false"
  }
}`),
			},
			wantErr: true,
		},
		{
			name: "TOML valid",
			args: args{
				loadedFlags: tomlFile,
			},
			flagFormat: "toml",
			expected: map[string]model.FlagData{
				"test-flag": {
					Rule:       testconvert.String("key eq \"random-key\""),
					Percentage: testconvert.Float64(100),
					True:       testconvert.Interface(true),
					False:      testconvert.Interface(false),
					Default:    testconvert.Interface(false),
				},
			},
			wantErr: false,
		},
		{
			name: "TOML invalid file",
			args: args{
				loadedFlags: []byte(`[test-flag]
rule = "key eq \"random-key\""
percentage = 100.0
true = true
false = false
default = false"
disable = false`),
			},
			flagFormat: "toml",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fCache := cache.New(cache.NewNotificationService([]notifier.Notifier{}))
			err := fCache.UpdateCache(tt.args.loadedFlags, tt.flagFormat)
			if tt.wantErr {
				assert.Error(t, err, "UpdateCache() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.NoError(t, err, "UpdateCache() error = %v, wantErr %v", err, tt.wantErr)
			// If no error we compare with expected
			for key, expected := range tt.expected {
				got, _ := fCache.GetFlag(key)
				assert.Equal(t, expected.GetRule(), got.GetRule())
				assert.Equal(t, expected.GetPercentage(), got.GetPercentage())
				assert.Equal(t, expected.GetTrue(), got.GetTrue())
				assert.Equal(t, expected.GetFalse(), got.GetFalse())
				assert.Equal(t, expected.GetDefault(), got.GetDefault())
				assert.Equal(t, expected.GetTrackEvents(), got.GetTrackEvents())
				assert.Equal(t, expected.GetDisable(), got.GetDisable())
				assert.Equal(t, expected.GetRollout(), got.GetRollout())
			}
			fCache.Close()
		})
	}
}
