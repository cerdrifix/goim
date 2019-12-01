-- CREATE ROLE cerdrifix WITH
--   LOGIN
--   PASSWORD 'cerdrifix1234'
--   NOSUPERUSER
--   INHERIT
--   NOCREATEDB
--   NOCREATEROLE
--   NOREPLICATION;


-- CREATE DATABASE goim;
-- alter database goim owner to cerdrifix;
-- 
-- Extension: "uuid-ossp"
-- CREATE EXTENSION IF NOT EXISTS "uuid-ossp"
--     SCHEMA public
--     VERSION '1.1';
	
-- Cleaning tables	
DROP TABLE IF EXISTS public.variables;
DROP TABLE IF EXISTS public.variables_type;
DROP TABLE IF EXISTS public.instances CASCADE;
DROP TABLE IF EXISTS public.states CASCADE;
DROP TABLE IF EXISTS public.users;
DROP TABLE IF EXISTS public.maps;
DROP TABLE IF EXISTS public.events;


-- Table: public.maps
CREATE TABLE IF NOT EXISTS public.maps
(
    id uuid NOT NULL DEFAULT uuid_generate_v1mc(),
    name varchar(255) COLLATE pg_catalog."default" NOT NULL,
    version integer NOT NULL,
    creation_date timestamp without time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    data json NOT NULL,
    CONSTRAINT maps_pkey PRIMARY KEY (id),
    CONSTRAINT maps_unique_name_version UNIQUE (name, version)

) TABLESPACE pg_default;

ALTER TABLE public.maps
    OWNER to cerdrifix;

-- Table: users
CREATE TABLE public.users
(
    username character varying(64) NOT NULL,
    name character varying(255) NOT NULL,
    surname character varying(255) NOT NULL,
    email character varying(255) NOT NULL,
    PRIMARY KEY (username)
);

ALTER TABLE public.users
    OWNER to cerdrifix;

-- Table: instances

CREATE TABLE public.instances
(
    id uuid NOT NULL DEFAULT uuid_generate_v1mc(),
    map_id uuid NOT NULL,
    start_date timestamp without time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    end_date timestamp without time zone,
    current_state uuid NULL,
    CONSTRAINT instances_pk PRIMARY KEY (id)
);

ALTER TABLE public.instances
    OWNER to cerdrifix;

-- Table: public.states

CREATE TABLE public.states
(
    id uuid NOT NULL DEFAULT uuid_generate_v1mc(),
    instance_id uuid NOT NULL,
    node_name character varying(255) COLLATE pg_catalog."default" NOT NULL,
    creator_id character varying(64) COLLATE pg_catalog."default" NOT NULL,
    owner_id character varying(64) COLLATE pg_catalog."default",
    enter_date timestamp without time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    leave_date timestamp without time zone,
    CONSTRAINT states_pk PRIMARY KEY (id)
) TABLESPACE pg_default;

ALTER TABLE public.states
    OWNER to cerdrifix;


-- Table: public.variables_type
CREATE TABLE public.variables_type
(
    name character varying(255) NOT NULL,
    type character varying(64) NOT NULL,
    CONSTRAINT variables_type_pk PRIMARY KEY (name)
);

ALTER TABLE public.variables_type
    OWNER to cerdrifix;


-- Table: public.variables
CREATE TABLE public.variables
(
    state_id uuid NOT NULL,
    name character varying(255) COLLATE pg_catalog."default" NOT NULL,
    data_type character varying(50) COLLATE pg_catalog."default" NOT NULL,
    data_value character varying COLLATE pg_catalog."default" NOT NULL,
    CONSTRAINT variables_pk PRIMARY KEY (state_id, name)
) TABLESPACE pg_default;

ALTER TABLE public.variables
    OWNER to cerdrifix;

-- Table: public.events
CREATE TABLE public.events
(
    event_id uuid NOT NULL DEFAULT uuid_generate_v1mc(),
    event_type varchar(128) NOT NULL,
    reference_id uuid NOT NULL,
    details json
) TABLESPACE  pg_default;

ALTER TABLE public.events
    OWNER TO cerdrifix;

------ Constraints
ALTER TABLE public.instances
ADD CONSTRAINT instances_fk_states FOREIGN KEY (current_state)
        REFERENCES public.states (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION;
		
ALTER TABLE public.instances
ADD	CONSTRAINT instances_fk_maps FOREIGN KEY (map_id)
        REFERENCES public.maps (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION;
		
ALTER TABLE public.states
ADD CONSTRAINT states_fk_instance_id FOREIGN KEY (instance_id)
        REFERENCES public.instances (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION;
		
ALTER TABLE public.states
ADD CONSTRAINT states_fk_users_creator FOREIGN KEY (creator_id)
        REFERENCES public.users (username) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION;
		
ALTER TABLE public.states
ADD CONSTRAINT states_fk_users_owner FOREIGN KEY (owner_id)
        REFERENCES public.users (username) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION;

ALTER TABLE public.variables
ADD CONSTRAINT variables_fk_states FOREIGN KEY (state_id)
        REFERENCES public.states (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION;

ALTER TABLE public.variables
ADD CONSTRAINT variables_fk_variables_type FOREIGN KEY (name)
        REFERENCES public.variables_type (name) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION;

------ Stored Procedures

-- Procedure sp_map_insert

CREATE OR REPLACE PROCEDURE public.sp_map_insert (
    _data json
)
AS
$$
DECLARE
    _name		VARCHAR(255);
    _version 	INTEGER;
BEGIN

    _name := _data->'name';
    _name := trim(both '"' from _name);

    if _name is null then
        raise exception 'Errore! JSON non contenente il nome del workflow';
    end if;

    SELECT 	COUNT(name) + 1 INTO _version
    FROM	public.maps
    WHERE	name = _name;

    INSERT INTO public.maps ( name, version, data)
    VALUES ( _name, _version, _data );

    raise notice 'Inserita mappa % - versione %', _name, _version;

END
$$ LANGUAGE plpgsql;

alter procedure public.sp_map_insert(json) owner to cerdrifix;

-- Procedure sp_maps_getlatestbyname
DROP FUNCTION IF EXISTS public.fn_maps_getlatestbyname(varchar);

CREATE FUNCTION public.fn_maps_getlatestbyname (
    _name varchar(255)
)
    RETURNS TABLE (
                      id		uuid,
                      name 	varchar(255),
                      version	int,
                      data	json
                  )
AS
$$
DECLARE
BEGIN

    RETURN QUERY
        SELECT 		M.id, M.name, M.version, M.data
        FROM		public.maps M
        WHERE		M.name = _name
        ORDER BY	version desc
        LIMIT 1;

END
$$ LANGUAGE plpgsql;

ALTER FUNCTION public.fn_maps_getlatestbyname(varchar) OWNER TO cerdrifix;

-- Procedure sp_maps_getbynameandversion
DROP FUNCTION IF EXISTS fn_maps_getbynameandversion;

CREATE OR REPLACE FUNCTION fn_maps_getbynameandversion (
    _name 		varchar(255),
    _version 	int
)
    RETURNS TABLE (
                      id		uuid,
                      name 	varchar(255),
                      version	int,
                      data	json
                  )
AS
$$
DECLARE
BEGIN

    RETURN QUERY
        SELECT 		M.id, M.name, M.version, M.data
        FROM		public.maps M
        WHERE		name = _name
        AND			version = _version;

END
$$ LANGUAGE plpgsql;

ALTER FUNCTION public.fn_maps_getbynameandversion(character varying, integer) OWNER TO cerdrifix;


-- Procedure sp_user_insert
CREATE OR REPLACE PROCEDURE sp_user_insert (
    username varchar(64),
	name varchar(255),
	surname varchar(255),
	email varchar(255)
)
AS
$$
BEGIN

	INSERT INTO public.users (username, name, surname, email)
	VALUES (username, name, surname, email);

    raise notice 'Inserito utente % - % %', username, name, surname;

END
$$ LANGUAGE plpgsql;

ALTER PROCEDURE public.sp_user_insert(varchar, varchar, varchar, varchar) OWNER TO cerdrifix;


-- Procedure fn_instance_new
CREATE OR REPLACE FUNCTION public.fn_instance_new(
	map_id uuid,
	start_node varchar(255),
	creator_id varchar(64),
	variables json
)
RETURNS uuid AS $inst_id$
DECLARE
	inst_id uuid;
	stat_id uuid;
	_key varchar(255);
	_value json;
	_type varchar(64);
BEGIN

	-- Creating instance
	INSERT INTO public.instances (map_id)
 	VALUES (map_id)
 	RETURNING id INTO inst_id;
	
	raise notice 'Instance id: %', inst_id;
	
	-- Creating state
	INSERT INTO public.states (instance_id, node_name, creator_id, owner_id)
	VALUES (inst_id, start_node, creator_id, creator_id)
 	RETURNING id INTO stat_id;
	
	raise notice 'State id: %', stat_id;
	
	-- Update instance with current state
	UPDATE 	public.instances
	SET		current_state = stat_id
	WHERE	id = inst_id;
	
	raise notice 'Instance % updated with current_state = %', inst_id, stat_id;
	
	-- Adding variables to state
    FOR _key, _value IN
       SELECT * FROM json_each(variables)
    LOOP
		SELECT 	type INTO _type
		FROM 	public.variables_type 
		WHERE 	name = _key;
		
		IF _type IS NULL THEN
			RAISE EXCEPTION 'Errore! La variabile % non e'' definita', _key;
		END IF;
		
		RAISE NOTICE 'Adding variable %: % (%)', _key, _value, _type;
	   
		INSERT INTO public.variables (state_id, name, data_type, data_value)
		VALUES (stat_id, _key, _type, _value);
	   
    END LOOP;
	
	RETURN inst_id;
	
END;
$inst_id$ LANGUAGE plpgsql;

ALTER FUNCTION public.fn_instance_new(uuid, character varying, character varying, json) OWNER TO cerdrifix;

-- notify_event

CREATE OR REPLACE FUNCTION notify_event() RETURNS TRIGGER AS $$
DECLARE
    data json;
    notification json;
BEGIN

    -- Convert the old or new row to JSON, based on the kind of action.
    -- Action = DELETE?             -> OLD row
    -- Action = INSERT or UPDATE?   -> NEW row
    IF (TG_OP = 'DELETE') THEN
        data = row_to_json(OLD);
    ELSE
        data = row_to_json(NEW);
    END IF;

    -- Contruct the notification as a JSON string.
    notification = json_build_object(
            'table',TG_TABLE_NAME,
            'action', TG_OP,
            'data', data);


    -- Execute pg_notify(channel, notification)
    PERFORM pg_notify('events',notification::text);

    -- Result is ignored since this is an AFTER trigger
    RETURN NULL;
END;

$$ LANGUAGE plpgsql;



DO $$
    DECLARE
        data json := '{"name":"richiesta_con_approvazione","description":"Richiesta con approvazione","nodes":[{"name":"start_1","description":"Inizio","type":"start","events":{"pre":[{"type":"validator","name":"CheckInputVariable","parameters":[{"name":"inputVariableName","type":"variable","value":"nome"}]}],"post":[{"type":"function","name":"CopyVariable","parameters":[{"name":"srcVariable","type":"variable","value":"nome"},{"name":"dstVariable","type":"variable","value":"NOMINATIVO"}]}]},"triggers":[{"name":"auto","after":{"unit":"seconds","value":0},"transaction":"start_to_task_approvativo"}],"transactions":[{"name":"start_to_task_approvativo","description":"Eseguito","to":"task_approvativo","events":{"pre":[],"post":[]}}]},{"name":"task_approvativo","description":"Task di Approvazione","type":"task","events":{"pre":[],"post":[]},"triggers":[{"name":"auto_approve","after":{"unit":"days","value":10},"transaction":"task_approvativo_cancel"}],"transactions":[{"name":"task_approvativo_ok","description":"Approva","visible":true,"to":"end_ok","events":{"pre":[],"post":[]}},{"name":"task_approvativo_ko","description":"Rifiuta","visible":true,"to":"end_ko","events":{"pre":[],"post":[]}},{"name":"task_approvativo_cancel","description":"Annulla","visible":false,"to":"end_canceled","events":{"pre":[],"post":[]}}]},{"name":"end_ok","type":"end","description":"Richiesta terminata con successo"},{"name":"end_ko","type":"end","description":"Richiesta rifiutata"},{"name":"end_canceled","type":"end","description":"Richiesta annullata da sistema (tempo massimo raggiunto)"}]}';
	BEGIN

        call sp_map_insert(data);
		
		call sp_user_insert('cerdrifix', 'Davide', 'Ceretto', 'ceretto.davide@gmail.com');
		
		INSERT INTO public.variables_type ( name, type )
		VALUES 	( 'nome', 'string' ),
				( 'cognome', 'string' ),
				( 'dataCreazione', 'datetime' ),
				( 'testoRichiesta', 'string' ),
				( 'NOMINATIVO', 'string' );

    END $$;

-- FUNCTION fn_notify_event
CREATE OR REPLACE FUNCTION fn_notify_event() RETURNS TRIGGER AS $$
    DECLARE
        data json;
        notification json;
    BEGIN
        data = row_to_json(NEW);

        notification = json_build_object(
            'table', tg_table_name,
            'action', tg_op,
            'data', data
        );

        PERFORM pg_notify('events', notification::text);

        RETURN NULL;
    END;
$$ LANGUAGE plpgsql;

------ Triggers
CREATE TRIGGER events_notify
    AFTER INSERT ON public.events
    FOR EACH ROW EXECUTE PROCEDURE fn_notify_event()