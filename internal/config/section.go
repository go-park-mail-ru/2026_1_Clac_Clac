package config

const (
	sectionConfigDefaultMaxQuantityTasks  = 100
	sectionConfigDefaultMinQuantityTasks  = 0
	sectionConfigDefaultmaxLenNameSection = 128
)

type SectionHandler struct {
	MaxQuantityTasks  int `mapstructure:"max_quantity_tasks"`
	MinQuantityTasks  int `mapstructure:"min_quantity_tasks"`
	MaxLenNameSection int `mapstructure:"max_len_name_section"`
}

type Section struct {
	Handler SectionHandler `mapstructure:"handler"`
}

func DefaultSectionConfig() Section {
	return Section{
		Handler: SectionHandler{
			MaxQuantityTasks:  sectionConfigDefaultMaxQuantityTasks,
			MinQuantityTasks:  sectionConfigDefaultMinQuantityTasks,
			MaxLenNameSection: sectionConfigDefaultmaxLenNameSection,
		},
	}
}
