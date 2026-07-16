package output

import (
	"io"

	"go.yaml.in/yaml/v3"
)

func renderYAML(w io.Writer, set RecordSet, opts RenderOptions) error {
	layout := layoutOf(opts)
	data := mapsForJSONYAML(set, layout)
	out, err := yaml.Marshal(data)
	if err != nil {
		return err
	}
	_, err = w.Write(out)
	return err
}
