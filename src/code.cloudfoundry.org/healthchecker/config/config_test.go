package config_test

import (
	"io/ioutil"
	"time"

	. "code.cloudfoundry.org/healthchecker/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {
	// var (
	// 	cfgInFile  config.Config
	// 	configFile *os.File
	// )
	var config *Config

	// JustBeforeEach(func() {
	// 	var err error
	// 	configFile, err = ioutil.TempFile("", "healthchecker.config")
	// 	Expect(err).NotTo(HaveOccurred())

	// 	cfgBytes, err := yaml.Marshal(cfgInFile)
	// 	Expect(err).NotTo(HaveOccurred())

	// 	_, err = configFile.Write(cfgBytes)
	// 	Expect(err).NotTo(HaveOccurred())
	// })

	// AfterEach(func() {
	// 	os.RemoveAll(configFile.Name())
	// })

	BeforeEach(func() {
		var err error
		config, err = DefaultConfig()
		Expect(err).ToNot(HaveOccurred())
	})

	Describe("Initialize", func() {
		Context("defaults", func() {
			It("sets log level from the DefaultConfig", func() {
				Expect(config.LogLevel).To(Equal("info"))
			})

			It("sets times from the DefaultConfig", func() {
				Expect(config.HealthCheckPollInterval).To(Equal(10 * time.Second))
				Expect(config.HealthCheckTimeout).To(Equal(5 * time.Second))
				Expect(config.StartupDelayBuffer).To(Equal(5 * time.Second))
			})
		})

		It("can add the non-default yaml keys for processing", func() {
			var b = []byte(`
component_name: kittens
failure_counter_file: "/var/vcap/data/component/some-file.count"
healthcheck_endpoint:
  host:     "some-host"
  port:     8888
  user:     "some-user"
  password: "some-password"
  path:     "/some-path"
start_response_delay_interval: 2s
log_level:                   "potato"
`)

			err := config.Initialize(b)
			Expect(err).ToNot(HaveOccurred())

			Expect(config.ComponentName).To(Equal("kittens"))
			Expect(config.FailureCounterFile).To(Equal("/var/vcap/data/component/some-file.count"))
			Expect(config.HealthCheckEndpoint.Host).To(Equal("some-host"))
			Expect(config.HealthCheckEndpoint.Port).To(Equal(8888))
			Expect(config.HealthCheckEndpoint.User).To(Equal("some-user"))
			Expect(config.HealthCheckEndpoint.Password).To(Equal("some-password"))
			Expect(config.HealthCheckEndpoint.Path).To(Equal("/some-path"))
			Expect(config.StartResponseDelayInterval).To(Equal(2 * time.Second))
			Expect(config.LogLevel).To(Equal("potato"))
		})
	})

	// Describe("LoadConfig", func() {
	// 	Context("when values specified in config file", func() {
	// 		BeforeEach(func() {
	// 			cfgInFile = config.Config{
	// 				ComponentName:      "healthchecker",
	// 				FailureCounterFile: "/var/vcap/data/component/some-file.count",
	// 				HealthCheckEndpoint: config.HealthCheckEndpoint{
	// 					Host:     "some-host",
	// 					Port:     8888,
	// 					User:     "some-user",
	// 					Password: "some-password",
	// 					Path:     "/some-path",
	// 				},
	// 				HealthCheckPollInterval:    3 * time.Minute,
	// 				HealthCheckTimeout:         4 * time.Hour,
	// 				StartupDelayBuffer:         1 * time.Millisecond,
	// 				StartResponseDelayInterval: 2 * time.Second,
	// 				LogLevel:                   "info",
	// 			}
	// 		})

	// 		It("loads values from the config file", func() {
	// 			c, err := config.LoadConfig(configFile.Name())
	// 			Expect(err).NotTo(HaveOccurred())
	// 			Expect(c).To(Equal(cfgInFile))
	// 		})

	// 		Context("when socket is provided", func() {
	// 			BeforeEach(func() {
	// 				cfgInFile.HealthCheckEndpoint.Socket = "/var/vcap/data/program/unix.sock"
	// 				cfgInFile.HealthCheckEndpoint.Host = ""
	// 				cfgInFile.HealthCheckEndpoint.Port = 0
	// 			})
	// 			Context("when host is provided", func() {
	// 				BeforeEach(func() {
	// 					cfgInFile.HealthCheckEndpoint.Host = "localhost"
	// 				})
	// 				It("throws an error when host is provided", func() {
	// 					_, err := config.LoadConfig(configFile.Name())
	// 					Expect(err).To(HaveOccurred())
	// 					Expect(err.Error()).To(ContainSubstring("Cannot specify both healthcheck endpoint host and socket"))

	// 				})
	// 			})
	// 			Context("when port is provided", func() {
	// 				BeforeEach(func() {
	// 					cfgInFile.HealthCheckEndpoint.Port = 1234
	// 				})
	// 				It("throws an error when port is provided", func() {
	// 					_, err := config.LoadConfig(configFile.Name())
	// 					Expect(err).To(HaveOccurred())
	// 					Expect(err.Error()).To(ContainSubstring("Cannot specify both healthcheck endpoint port and socket"))
	// 				})
	// 			})
	// 			It("loads values from the config file", func() {
	// 				c, err := config.LoadConfig(configFile.Name())
	// 				Expect(err).NotTo(HaveOccurred())
	// 				Expect(c).To(Equal(cfgInFile))

	// 			})
	// 			Context("when StartResponseDelayInteral is provided", func() {
	// 				BeforeEach(func() {
	// 					cfgInFile.StartResponseDelayInterval = 3 * time.Second
	// 				})
	// 				It("throws an error when ", func() {
	// 					c, err := config.LoadConfig(configFile.Name())
	// 					Expect(err).ToNot(HaveOccurred())
	// 					Expect(c.StartResponseDelayInterval).To(Equal(3 * time.Second))
	// 				})

	// 			})
	// 		})
	// 	})

	// 	Context("when required properties are not provided", func() {
	// 		BeforeEach(func() {
	// 			cfgInFile = config.Config{
	// 				ComponentName: "healthchecker",
	// 				HealthCheckEndpoint: config.HealthCheckEndpoint{
	// 					Host:     "some-host",
	// 					Port:     8888,
	// 					User:     "some-user",
	// 					Password: "some-password",
	// 					Path:     "/some-path",
	// 				},
	// 			}
	// 		})

	// 		Context("when component_name is empty", func() {
	// 			BeforeEach(func() {
	// 				cfgInFile.ComponentName = ""
	// 			})

	// 			It("returns an error", func() {
	// 				_, err := config.LoadConfig(configFile.Name())
	// 				Expect(err).To(HaveOccurred())
	// 				Expect(err.Error()).To(ContainSubstring("Missing component_name"))
	// 			})
	// 		})

	// 		Context("when failure_counter_file is empty", func() {
	// 			BeforeEach(func() {
	// 				cfgInFile.FailureCounterFile = ""
	// 			})

	// 			It("returns an error", func() {
	// 				_, err := config.LoadConfig(configFile.Name())
	// 				Expect(err).To(HaveOccurred())
	// 				Expect(err.Error()).To(ContainSubstring("Missing failure_counter_file"))
	// 			})
	// 		})

	// 		Context("when socket is not provided", func() {
	// 			Context("when host is empty", func() {
	// 				BeforeEach(func() {
	// 					cfgInFile.HealthCheckEndpoint.Host = ""
	// 				})

	// 				It("returns an error", func() {
	// 					_, err := config.LoadConfig(configFile.Name())
	// 					Expect(err).To(HaveOccurred())
	// 					Expect(err.Error()).To(ContainSubstring("Missing healthcheck endpoint host or socket"))
	// 				})
	// 			})

	// 			Context("when port is empty", func() {
	// 				BeforeEach(func() {
	// 					cfgInFile.HealthCheckEndpoint.Port = 0
	// 				})

	// 				It("returns an error", func() {
	// 					_, err := config.LoadConfig(configFile.Name())
	// 					Expect(err).To(HaveOccurred())
	// 					Expect(err.Error()).To(ContainSubstring("Missing healthcheck endpoint port or socket"))
	// 				})
	// 			})
	// 		})
	// 	})

	// 	Context("when values with defaults are not provided", func() {
	// 		BeforeEach(func() {
	// 			cfgInFile = config.Config{
	// 				ComponentName:      "healthchecker",
	// 				FailureCounterFile: "/var/vcap/data/component/some-file.count",
	// 				HealthCheckEndpoint: config.HealthCheckEndpoint{
	// 					Host:     "some-host",
	// 					Port:     8888,
	// 					User:     "some-user",
	// 					Password: "some-password",
	// 					Path:     "/some-path",
	// 				},
	// 			}
	// 		})

	// 		It("sets default values when they are not provided", func() {
	// 			c, err := config.LoadConfig(configFile.Name())
	// 			Expect(err).NotTo(HaveOccurred())
	// 			Expect(c.HealthCheckPollInterval).To(Equal(config.DefaultConfig.HealthCheckPollInterval))
	// 			Expect(c.HealthCheckTimeout).To(Equal(config.DefaultConfig.HealthCheckTimeout))
	// 			Expect(c.StartResponseDelayInterval).To(Equal(config.DefaultConfig.StartResponseDelayInterval))
	// 			Expect(c.StartupDelayBuffer).To(Equal(config.DefaultConfig.StartupDelayBuffer))
	// 			Expect(c.LogLevel).To(Equal(config.DefaultConfig.LogLevel))
	// 		})
	// 	})

	Context("when file does not exist", func() {
		It("returns an error", func() {
			_, err := LoadConfig("does-not-exist")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Could not read config file"))
		})
	})

	Context("when config file is malformed", func() {
		It("returns an error", func() {
			configFile, err := ioutil.TempFile("", "healthchecker-malformed.config")
			Expect(err).NotTo(HaveOccurred())

			_, err = configFile.WriteString("meow")
			Expect(err).NotTo(HaveOccurred())

			_, err = LoadConfig(configFile.Name())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Could not unmarshal config file"))
		})
	})
})
