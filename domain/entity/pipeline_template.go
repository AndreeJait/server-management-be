package entity

type PipelineTemplateDefinition struct {
	Steps []PipelineStepDefinition
}

type PipelineStepDefinition struct {
	Name   string
	Order  int
	Config string
}

func DefaultPipelineTemplate(preset FrameworkPreset) *PipelineTemplateDefinition {
	baseSteps := []PipelineStepDefinition{
		{Name: "pull_image", Order: 1, Config: `{"force":false}`},
		{Name: "create_container", Order: 2, Config: `{"port":8080,"env":{},"volume_mounts":[],"healthcheck_path":"/"}`},
		{Name: "start_container", Order: 3, Config: `{}`},
		{Name: "verify_health", Order: 4, Config: `{"timeout_seconds":30,"interval_seconds":2}`},
	}

	switch preset {
	case FrameworkNextjs, FrameworkNuxt, FrameworkAstro:
		baseSteps[1].Config = `{"port":3000,"env":{"NODE_ENV":"production"},"volume_mounts":[],"healthcheck_path":"/"}`
	case FrameworkReact, FrameworkVue, FrameworkSvelte, FrameworkRemix:
		baseSteps[1].Config = `{"port":3000,"env":{"NODE_ENV":"production"},"volume_mounts":[],"healthcheck_path":"/"}`
	case FrameworkGo:
		baseSteps[1].Config = `{"port":8080,"env":{},"volume_mounts":[],"healthcheck_path":"/health"}`
	case FrameworkPython, FrameworkNodejs:
		baseSteps[1].Config = `{"port":8000,"env":{},"volume_mounts":[],"healthcheck_path":"/"}`
	case FrameworkJava:
		baseSteps[1].Config = `{"port":8080,"env":{},"volume_mounts":[],"healthcheck_path":"/actuator/health"}`
	case FrameworkRust:
		baseSteps[1].Config = `{"port":8080,"env":{},"volume_mounts":[],"healthcheck_path":"/"}`
	case FrameworkDotnet:
		baseSteps[1].Config = `{"port":8080,"env":{},"volume_mounts":[],"healthcheck_path":"/health"}`
	case FrameworkLaravel:
		baseSteps[1].Config = `{"port":8000,"env":{},"volume_mounts":[],"healthcheck_path":"/"}`
		baseSteps = append(baseSteps,
			PipelineStepDefinition{Name: "exec_command", Order: 5, Config: `{"command":"php artisan storage:link"}`},
			PipelineStepDefinition{Name: "exec_command", Order: 6, Config: `{"command":"php artisan migrate --force"}`},
			PipelineStepDefinition{Name: "exec_command", Order: 7, Config: `{"command":"php artisan config:cache"}`},
			PipelineStepDefinition{Name: "exec_command", Order: 8, Config: `{"command":"php artisan view:cache"}`},
		)
	case FrameworkRails:
		baseSteps[1].Config = `{"port":8000,"env":{},"volume_mounts":[],"healthcheck_path":"/"}`
		baseSteps = append(baseSteps,
			PipelineStepDefinition{Name: "exec_command", Order: 5, Config: `{"command":"rails db:migrate"}`},
		)
	}

	return &PipelineTemplateDefinition{Steps: baseSteps}
}