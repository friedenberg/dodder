package env_vars

type EnvVars map[string]string

type Adder interface {
	AddToEnvVars(EnvVars)
}

func Make(adders ...Adder) EnvVars {
	envVars := make(EnvVars)

	for _, adder := range adders {
		adder.AddToEnvVars(envVars)
	}

	return envVars
}
