CREATE TABLE `data` (
	`id` INT(11) NOT NULL AUTO_INCREMENT,
	`time` TIMESTAMP NOT NULL DEFAULT current_timestamp(),
	`DevBat` INT(11) NULL DEFAULT NULL COMMENT 'Device Battery percentage',
	`Gids` INT(11) NULL DEFAULT NULL,
	`Lat` FLOAT NULL DEFAULT NULL,
	`Long` FLOAT NULL DEFAULT NULL,
	`Elv` INT(11) NULL DEFAULT NULL COMMENT 'Elevation in meters',
	`Seq` INT(11) NULL DEFAULT NULL COMMENT 'Sequence number of transfer',
	`Trip` INT(10) UNSIGNED NULL DEFAULT NULL COMMENT 'Trip number',
	`odo` FLOAT UNSIGNED NULL DEFAULT NULL COMMENT 'Odometer in km',
	`SOC` FLOAT NULL DEFAULT NULL COMMENT 'State of charge',
	`AHr` FLOAT NULL DEFAULT NULL COMMENT 'Current AHr capacity',
	`BatTemp` FLOAT NULL DEFAULT NULL COMMENT 'Average battery temperature',
	`Amb` FLOAT NULL DEFAULT NULL COMMENT 'Ambient temperature',
	`Wpr` INT(11) NULL DEFAULT NULL COMMENT 'Front wiper status',
	`PlugState` INT(11) NULL DEFAULT NULL COMMENT 'Plug state',
	`ChrgMode` INT(11) NULL DEFAULT NULL COMMENT 'Charge mode',
	`ChrgPwr` INT(11) NULL DEFAULT NULL COMMENT 'Charge power in watts',
	`VIN` VARCHAR(20) NULL DEFAULT NULL,
	`PwrSw` INT(10) UNSIGNED NULL DEFAULT NULL COMMENT 'Power switch state',
	`Tunits` VARCHAR(1) NULL DEFAULT NULL COMMENT 'Temperature units',
	`RPM` INT(10) UNSIGNED NULL DEFAULT NULL COMMENT 'Motor RPM',
	`SOH` FLOAT UNSIGNED NULL DEFAULT NULL COMMENT 'State of health',
	PRIMARY KEY (`id`)
)
COLLATE='utf8mb4_general_ci'
ENGINE=InnoDB
AUTO_INCREMENT=1
;