package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseMap(t *testing.T) {
	input := ""
	expected := map[string]string{}
	res, err := parseMap(input)
	assert.NoError(t, err)
	assert.Equal(t, expected, res)
}

func TestParseMapErr(t *testing.T) {
	input := "asdads"
	expected := map[string]string{}
	res, err := parseMap(input)
	assert.Error(t, err)
	assert.Equal(t, expected, res)
}

func TestParseMapErr1(t *testing.T) {
	input := "asdads=asda,ttt"
	expected := map[string]string{}
	res, err := parseMap(input)
	assert.Error(t, err)
	assert.Equal(t, expected, res)
}

func TestParseMapOk(t *testing.T) {
	input := "asdads=asda,ttt=123"
	expected := map[string]string{
		"asdads": "asda",
		"ttt":    "123",
	}
	res, err := parseMap(input)
	assert.NoError(t, err)
	assert.Equal(t, expected, res)
}

func TestValidate(t *testing.T) {

}
