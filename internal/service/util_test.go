package service_test

import (
	"testing"

	"com.telekom-mms.corp-net-indicator/internal/service"
	"com.telekom-mms.corp-net-indicator/internal/test"
	"github.com/godbus/dbus/v5"
	"github.com/stretchr/testify/assert"
)

type DeepStruct struct {
	Deep int
}

type TestStruct struct {
	FieldA string
	FieldB int
	FieldC bool
}

type TestStructB struct {
	FieldA *string
	FieldB *int
	FieldC *bool
}

func TestMapDbusDictToStruct(t *testing.T) {
	assert := assert.New(t)

	// good case
	assert.Equal(&TestStruct{
		FieldA: "hu",
		FieldB: 1,
		FieldC: true,
	}, service.MapDbusDictToStruct(map[string]dbus.Variant{
		"FieldA": dbus.MakeVariant("hu"),
		"FieldB": dbus.MakeVariant(1),
		"FieldC": dbus.MakeVariant(true),
	}, &TestStruct{}))

	// missing variant
	assert.Equal(&TestStruct{
		FieldA: "hu",
		FieldB: 1,
		FieldC: false,
	}, service.MapDbusDictToStruct(map[string]dbus.Variant{
		"FieldA": dbus.MakeVariant("hu"),
		"FieldB": dbus.MakeVariant(1),
	}, &TestStruct{}))

	// to many variants
	assert.Equal(&TestStruct{
		FieldA: "hu",
		FieldB: 1,
		FieldC: false,
	}, service.MapDbusDictToStruct(map[string]dbus.Variant{
		"FieldA": dbus.MakeVariant("hu"),
		"FieldB": dbus.MakeVariant(1),
		"FieldE": dbus.MakeVariant(3.4),
	}, &TestStruct{}))

	// wrong types
	assert.Equal(&TestStruct{
		FieldA: "",
		FieldB: 0,
		FieldC: false,
	}, service.MapDbusDictToStruct(map[string]dbus.Variant{
		"FieldA": dbus.MakeVariant(1000),
		"FieldB": dbus.MakeVariant(false),
	}, &TestStruct{}))
}

func TestMapDbusDictToStructWithPointer(t *testing.T) {
	assert := assert.New(t)

	// good case
	assert.Equal(&TestStructB{
		FieldA: test.Pointer("hu"),
		FieldB: test.Pointer(1),
		FieldC: test.Pointer(true),
	}, service.MapDbusDictToStruct(map[string]dbus.Variant{
		"FieldA": dbus.MakeVariant("hu"),
		"FieldB": dbus.MakeVariant(1),
		"FieldC": dbus.MakeVariant(true),
	}, &TestStructB{}))

	// missing variant
	assert.Equal(&TestStructB{
		FieldA: test.Pointer("hu"),
		FieldB: test.Pointer(1),
		FieldC: nil,
	}, service.MapDbusDictToStruct(map[string]dbus.Variant{
		"FieldA": dbus.MakeVariant("hu"),
		"FieldB": dbus.MakeVariant(1),
	}, &TestStructB{}))

	// to many variants
	assert.Equal(&TestStructB{
		FieldA: test.Pointer("hu"),
		FieldB: test.Pointer(1),
		FieldC: nil,
	}, service.MapDbusDictToStruct(map[string]dbus.Variant{
		"FieldA": dbus.MakeVariant("hu"),
		"FieldB": dbus.MakeVariant(1),
		"FieldE": dbus.MakeVariant(3.4),
	}, &TestStructB{}))

	// wrong types
	assert.Equal(&TestStructB{
		FieldA: nil,
		FieldB: nil,
		FieldC: nil,
	}, service.MapDbusDictToStruct(map[string]dbus.Variant{
		"FieldA": dbus.MakeVariant(1000),
		"FieldB": dbus.MakeVariant(false),
	}, &TestStructB{}))
}
