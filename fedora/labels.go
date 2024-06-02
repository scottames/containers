package main

// TODO: docs
type ContainerLabel struct {
	Name  string
	Value string
}

// WithLabel will append a label to the output container
//
// maps not currently supported: https://github.com/dagger/dagger/issues/6138
// if they do become supported this should be simplified to allow passing a
// map
func (f *Fedora) WithLabel(label struct {
	//+optional
	Name *string
	//+optional
	Value *string
},
) *Fedora {
	if label.Name != nil && label.Value != nil {
		f.Labels = append(f.Labels, &ContainerLabel{
			Name:  *label.Name,
			Value: *label.Value,
		})
	}

	return f
}

// TODO: docs
func (f *Fedora) WithDescription(description string) *Fedora {
	f.Labels = append(f.Labels, &ContainerLabel{
		Name:  "org.opencontainers.image.description",
		Value: description,
	})

	return f
}
