package builder

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/SUSE/fissile/scripts/configgin"
	"github.com/SUSE/fissile/scripts/dockerfiles"
	"github.com/SUSE/fissile/util"
)

// BaseImageBuilder represents a builder of docker base images
type BaseImageBuilder struct {
	BaseImage string
}

// NewBaseImageBuilder creates a new BaseImageBuilder
func NewBaseImageBuilder(baseImage string) *BaseImageBuilder {
	return &BaseImageBuilder{
		BaseImage: baseImage,
	}
}

// NewDockerPopulator returns a function that will populate the docker tar archive
func (b *BaseImageBuilder) NewDockerPopulator() func(*tar.Writer) error {
	return func(tarWriter *tar.Writer) error {
		// Generate dockerfile
		dockerfileContents, err := b.generateDockerfile()
		if err != nil {
			return err
		}
		err = util.WriteToTarStream(tarWriter, dockerfileContents, tar.Header{
			Name: "Dockerfile",
		})
		if err != nil {
			return err
		}

		// Add rsyslog_conf, monitrc.erb, and the post-start handler.
		for _, assetName := range dockerfiles.AssetNames() {
			switch {
			case strings.HasPrefix(assetName, "rsyslog_conf/"):
			case assetName == "monitrc.erb":
			case assetName == "post-start.sh":
			default:
				continue
			}
			assetContents, err := dockerfiles.Asset(assetName)
			if err != nil {
				return err
			}
			err = util.WriteToTarStream(tarWriter, assetContents, tar.Header{
				Name: assetName,
			})
			if err != nil {
				return err
			}
		}

		// Add configgin
		configginGzip, err := configgin.Asset("configgin.tgz")
		if err != nil {
			return err
		}
		err = util.TargzIterate(
			"configgin.tgz",
			bytes.NewReader(configginGzip),
			func(reader *tar.Reader, header *tar.Header) error {
				header.Name = filepath.Join("configgin", header.Name)
				if err = tarWriter.WriteHeader(header); err != nil {
					return err
				}
				if _, err = io.Copy(tarWriter, reader); err != nil {
					return err
				}
				return nil
			})
		if err != nil {
			return err
		}

		return nil
	}
}

func (b *BaseImageBuilder) generateDockerfile() ([]byte, error) {
	asset, err := dockerfiles.Asset("Dockerfile-base")
	if err != nil {
		return nil, err
	}

	dockerfileTemplate := template.New("Dockerfile-base")
	dockerfileTemplate, err = dockerfileTemplate.Parse(string(asset))
	if err != nil {
		return nil, err
	}

	var output bytes.Buffer
	err = dockerfileTemplate.Execute(&output, b)
	if err != nil {
		return nil, err
	}

	return output.Bytes(), nil
}

// GetBaseImageName generates a docker image name to be used as a role image base
func GetBaseImageName(repository, fissileVersion string) string {
	return util.SanitizeDockerName(fmt.Sprintf("%s-role-base:%s", repository, fissileVersion))
}
