package config_test

import (
	"io/ioutil"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v2"

	"code.cloudfoundry.org/cf-networking-helpers/healthchecker/config"
)

var _ = Describe("Config", func() {
	var (
		cfgInFile  config.Config
		configFile *os.File
	)

	JustBeforeEach(func() {
		var err error
		configFile, err = ioutil.TempFile("", "healthchecker.config")
		Expect(err).NotTo(HaveOccurred())

		cfgBytes, err := yaml.Marshal(cfgInFile)
		Expect(err).NotTo(HaveOccurred())

		_, err = configFile.Write(cfgBytes)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		os.RemoveAll(configFile.Name())
	})

	Describe("LoadConfig", func() {
		Context("when values specified in config file", func() {
			BeforeEach(func() {
				cfgInFile = config.Config{
					ComponentName:      "healthchecker",
					FailureCounterFile: "/var/vcap/data/component/some-file.count",
					HealthCheckEndpoint: config.HealthCheckEndpoint{
						Host:     "some-host",
						Port:     8888,
						User:     "some-user",
						Password: "some-password",
						Path:     "/some-path",
					},
					HealthCheckPollInterval:    3 * time.Minute,
					HealthCheckTimeout:         4 * time.Hour,
					StartupDelayBuffer:         1 * time.Millisecond,
					StartResponseDelayInterval: 2 * time.Second,
					LogLevel:                   "info",
				}
			})

			It("loads values from the config file", func() {
				c, err := config.LoadConfig(configFile.Name())
				Expect(err).NotTo(HaveOccurred())
				Expect(c).To(Equal(cfgInFile))
			})

			Context("when socket is provided", func() {
				BeforeEach(func() {
					cfgInFile.HealthCheckEndpoint.Socket = "/var/vcap/data/program/unix.sock"
					cfgInFile.HealthCheckEndpoint.Host = ""
					cfgInFile.HealthCheckEndpoint.Port = 0
				})
				Context("when host is provided", func() {
					BeforeEach(func() {
						cfgInFile.HealthCheckEndpoint.Host = "localhost"
					})
					It("throws an error when host is provided", func() {
						_, err := config.LoadConfig(configFile.Name())
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("Cannot specify both healthcheck endpoint host and socket"))

					})
				})
				Context("when port is provided", func() {
					BeforeEach(func() {
						cfgInFile.HealthCheckEndpoint.Port = 1234
					})
					It("throws an error when port is provided", func() {
						_, err := config.LoadConfig(configFile.Name())
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("Cannot specify both healthcheck endpoint port and socket"))
					})
				})
				It("loads values from the config file", func() {
					c, err := config.LoadConfig(configFile.Name())
					Expect(err).NotTo(HaveOccurred())
					Expect(c).To(Equal(cfgInFile))

				})
			})
		})

		Context("when required properties are not provided", func() {
			BeforeEach(func() {
				cfgInFile = config.Config{
					ComponentName: "healthchecker",
					HealthCheckEndpoint: config.HealthCheckEndpoint{
						Host:     "some-host",
						Port:     8888,
						User:     "some-user",
						Password: "some-password",
						Path:     "/some-path",
					},
				}
			})

			Context("when component_name is empty", func() {
				BeforeEach(func() {
					cfgInFile.ComponentName = ""
				})

				It("returns an error", func() {
					_, err := config.LoadConfig(configFile.Name())
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Missing component_name"))
				})
			})

			Context("when failure_counter_file is empty", func() {
				BeforeEach(func() {
					cfgInFile.FailureCounterFile = ""
				})

				It("returns an error", func() {
					_, err := config.LoadConfig(configFile.Name())
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Missing failure_counter_file"))
				})
			})

			Context("when socket is not provided", func() {
				Context("when host is empty", func() {
					BeforeEach(func() {
						cfgInFile.HealthCheckEndpoint.Host = ""
					})

					It("returns an error", func() {
						_, err := config.LoadConfig(configFile.Name())
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("Missing healthcheck endpoint host or socket"))
					})
				})

				Context("when port is empty", func() {
					BeforeEach(func() {
						cfgInFile.HealthCheckEndpoint.Port = 0
					})

					It("returns an error", func() {
						_, err := config.LoadConfig(configFile.Name())
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("Missing healthcheck endpoint port or socket"))
					})
				})
			})
		})

		Context("when values with defaults are not provided", func() {
			BeforeEach(func() {
				cfgInFile = config.Config{
					ComponentName:      "healthchecker",
					FailureCounterFile: "/var/vcap/data/component/some-file.count",
					HealthCheckEndpoint: config.HealthCheckEndpoint{
						Host:     "some-host",
						Port:     8888,
						User:     "some-user",
						Password: "some-password",
						Path:     "/some-path",
					},
				}
			})

			It("sets default values when they are not provided", func() {
				c, err := config.LoadConfig(configFile.Name())
				Expect(err).NotTo(HaveOccurred())
				Expect(c.HealthCheckPollInterval).To(Equal(config.DefaultConfig.HealthCheckPollInterval))
				Expect(c.HealthCheckTimeout).To(Equal(config.DefaultConfig.HealthCheckTimeout))
				Expect(c.StartResponseDelayInterval).To(Equal(config.DefaultConfig.StartResponseDelayInterval))
				Expect(c.StartupDelayBuffer).To(Equal(config.DefaultConfig.StartupDelayBuffer))
				Expect(c.LogLevel).To(Equal(config.DefaultConfig.LogLevel))
			})
		})

		Context("when file does not exist", func() {
			It("returns an error", func() {
				_, err := config.LoadConfig("does-not-exist")
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

				_, err = config.LoadConfig(configFile.Name())
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Could not unmarshal config file"))
			})
		})
	})
})
