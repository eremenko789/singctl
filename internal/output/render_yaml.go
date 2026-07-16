package output

import (
	"io"

	"go.yaml.in/yaml/v3"
)

func renderYAML(w io.Writer, set RecordSet, opts RenderOptions) error {
	layout := layoutOf(opts)
	data := mapsForJSONYAML(set, layout)
	var payload any = data
	if opts.SingleObject {
		payload = data[0]
	}
	out, err := yaml.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = w.Write(out)
	return err
}
