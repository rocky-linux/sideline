//go:generate protoc -Iproto -I. --go_out=paths=source_relative:pb proto/configuration.proto
package sideline
