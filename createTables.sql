BEGIN;
	CREATE TABLE contract (
		contract TEXT NOT NULL, 
		duration TEXT NOT NULL, 
		version TEXT NOT NULL, 
		CONSTRAINT contract_pk PRIMARY KEY ("contract")
	);

	CREATE TABLE machine_sensor(
		id BIGSERIAL, machine TEXT NOT NULL, 
		sensor TEXT NOT NULL, 
		CONSTRAINT machie_sensor_pk PRIMARY KEY ("id"), 
		CONSTRAINT uniqueness UNIQUE("machine", "sensor")
	);

	CREATE TABLE contract_machine_sensor(
		contract TEXT NOT NULL, 
		machine_sensor BIGINT NOT NULL, 
		CONSTRAINT contract_machine_sensor_contract_fk FOREIGN KEY ("contract") REFERENCES contract(contract) ON DELETE CASCADE,
		CONSTRAINT contract_machine_sensor_machine_sensor_fk FOREIGN KEY ("machine_sensor") REFERENCES machine_sensor(id)
	);
COMMIT;
