package engine

import (
	"../routines"
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"log"
	"sync"
	"time"
)

type ProcessEngine struct {
	logger  *log.Logger
	db      *sqlx.DB
	context context.Context
}

func New(logger *log.Logger, db *sqlx.DB, context context.Context) *ProcessEngine {
	return &ProcessEngine{
		logger:  logger,
		db:      db,
		context: context,
	}
}

func (engine *ProcessEngine) GetProcessMap(name string) (processMap ProcessMap, err error) {

	q := fmt.Sprintf("SELECT * FROM public.fn_maps_getlatestbyname('%s')", name)

	rows, err := engine.db.QueryxContext(engine.context, q)
	if err != nil {
		engine.logger.Printf("Unable to query db")
		return processMap, err
	}
	for rows.Next() {
		var id uuid.UUID
		var name string
		var version int
		var data string

		err = rows.Scan(&id, &name, &version, &data)
		if err != nil {
			engine.logger.Printf("Error during row scan")
			return processMap, err
		}

		engine.logger.Printf("\nMap id: %s\n    name: %s\n    version: %d\n    data: %s", id, name, version, data)

		err = json.Unmarshal([]byte(data), &processMap)
		if err != nil {
			engine.logger.Printf("Error during unmarshaling \njson: %s\nerror: %v", data, err)
			return processMap, err
		}

		processMap.Id = id
	}
	return processMap, nil
}

func transformVariablesInJSONPayload(variables map[string]interface{}) (string, error) {

	fmt.Printf("Variables: %#v", variables)

	for k, v := range variables {

		switch t := v.(type) {
		case string:
			dt, err := time.Parse("2006-01-02T15:04:05.000", v.(string))
			if err == nil {
				variables[k] = dt
				fmt.Println(k, dt, "(datetime)")
			} else {
				variables[k] = v
				fmt.Println(k, v, "(string)")
			}
		case float64:
			variables[k] = v.(float64)
			fmt.Println(k, v, "(float64)")
		case int:
			variables[k] = v.(int)
			fmt.Println(k, v, "(int)")
		case []interface{}:
			fmt.Println(k, "(array):")
			for i, u := range t {
				fmt.Println("    ", i, u)
			}
		default:
			fmt.Println(k, v, "(unknown)")
		}
	}

	j, err := json.Marshal(variables)

	return string(j), err
}

func doProcessEvent(e ProcessEvent, variables *map[string]interface{}, engine *ProcessEngine, waitGroup *sync.WaitGroup) error {
	fmt.Printf("\n\nProcessing event: %#v", e)

	r := routines.New(engine.logger, variables)
	params := make([]interface{}, len(e.Parameters))

	for i, p := range e.Parameters {
		params[i] = p.Value
	}

	switch e.EventType {
	case "function":
		_, err := r.CallFunc(e.Name, params)

		if err != nil {
			return err
		}
	case "validator":
		_, err := r.CallFunc(e.Name, params)

		if err != nil {
			return err
		}
	}
	waitGroup.Done()
	return nil
}

func (engine *ProcessEngine) raiseError(message string, err error) []error {
	tErr := fmt.Errorf("%s: %v", message, err)
	engine.logger.Printf(tErr.Error())
	errs := make([]error, 1)
	errs[0] = err
	return errs
}

func (engine *ProcessEngine) NewInstance(payload CreateProcessPayload) (instanceNumber uuid.UUID, errs []error) {

	tx, err := engine.db.Beginx()
	if err != nil {
		return uuid.Nil, engine.raiseError("Error beginning transaction", err)
	}

	pmap, err := engine.GetProcessMap(payload.ProcessName)
	if err != nil {
		tx.Rollback()
		return uuid.Nil, engine.raiseError("Error retrieving map from database", err)
	}
	fmt.Printf("Map: %#v", pmap)

	engine.logger.Printf("Map: %s", pmap)

	var startNode ProcessNode
	var found bool

	for _, n := range pmap.Nodes {
		if n.NodeType == "start" {
			startNode = n
			found = true
			break
		}
	}

	if found == false {
		tx.Rollback()
		return uuid.Nil, engine.raiseError("Error retrieving start node", nil)
	}

	engine.logger.Printf("Start node found! %#v", startNode)

	// Processing pre events
	errors := engine.executeEvents(startNode.Events.Pre, &payload.Variables, tx)
	if len(errors) > 0 {
		return uuid.Nil, errors
	}

	j, err := transformVariablesInJSONPayload(payload.Variables)

	go engine.startTriggers(startNode.Triggers, &payload.Variables)

	fmt.Printf("Creating instance with:\n"+
		"  map_id: %s\n"+
		"  start_node: %s\n"+
		"  variables: %s",
		pmap.Id, startNode.Name, string(j))

	creator := "cerdrifix"

	q := fmt.Sprintf("SELECT public.fn_instance_new('%s', '%s', '%s', '%s')", pmap.Id, startNode.Name, creator, string(j))

	engine.logger.Printf("q: %s", q)

	err = engine.db.Get(&instanceNumber, q)
	if err != nil {
		tx.Rollback()
		return uuid.Nil, engine.raiseError("Errors occurred during instance creation", err)
	}

	tx.Commit()

	return instanceNumber, nil
}

func (engine *ProcessEngine) startTriggers(triggers []ProcessTrigger, variables *map[string]interface{}) {
	for _, trigger := range triggers {
		engine.logger.Printf("Trigger: %#v", trigger)
	}
}

func (engine *ProcessEngine) executeEvents(events []ProcessEvent, variables *map[string]interface{}, tx *sqlx.Tx) (errors []error) {
	var eventsWG sync.WaitGroup
	var err error
	eventsWG.Add(len(events))
	errors = make([]error, 0, len(events))
	for _, ev := range events {
		go func() {
			err = doProcessEvent(ev, variables, engine, &eventsWG)
			if err != nil {
				_ = append(errors, err)
			}
		}()
	}
	eventsWG.Wait()
	if len(errors) > 0 {
		tx.Rollback()
		engine.logger.Fatalf("Errors occured during pre-events: %#v", errors)
		return errors
	}
	return nil
}
