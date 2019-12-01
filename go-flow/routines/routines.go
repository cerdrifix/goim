package routines

import (
	"fmt"
	"log"
	"reflect"
)

type Routines struct {
	logger    *log.Logger
	variables *map[string]interface{}
}

func New(logger *log.Logger, variables *map[string]interface{}) *Routines {
	return &Routines{
		logger:    logger,
		variables: variables,
	}
}

func (r *Routines) CallFunc(name string, params []interface{}) (result []reflect.Value, err error) {

	r.logger.Printf("Called reflection with name: %s and params: %#v", name, params)

	f := reflect.ValueOf(r).MethodByName(name)

	r.logger.Printf("Function with name %s found by reflection: %#v", name, f)

	if len(params) != f.Type().NumIn() {
		err = fmt.Errorf("wrong number of params.\nfound: %d\nexpected: %d\nobject: %#v", len(params), f.Type().NumIn(), params)
		r.logger.Printf("Error!: %s", err)
		return
	}
	in := make([]reflect.Value, len(params))
	for k, param := range params {
		in[k] = reflect.ValueOf(param)
	}

	resErr := f.Call(in)[0].Interface()

	if resErr != nil {
		return result, resErr.(error)
	}

	return
}

func (r *Routines) CopyVariable(srcVariable string, dstVariable string) error {
	r.logger.Printf("Called CopyVariables(\"%s\",\"%s\")", srcVariable, dstVariable)
	variables := *r.variables

	if variables[srcVariable] == nil {
		var e string = fmt.Sprintf("Error occured processing routine CopyVariable: cannot fine variable %s", srcVariable)
		r.logger.Printf(e)
		return fmt.Errorf(e)
	}
	variables[dstVariable] = variables[srcVariable]

	r.logger.Printf("Variable %s successfully copied in %s", srcVariable, dstVariable)

	return nil
}

func (r *Routines) CheckInputVariable(name string) error {
	r.logger.Printf("Called CheckInputVariable(\"%s\")", name)
	if (*r.variables)[name] != nil {
		r.logger.Printf("Variable %s found!", name)
		return nil
	}
	return fmt.Errorf(fmt.Sprintf("Error! Variable %s not found", name))
}
