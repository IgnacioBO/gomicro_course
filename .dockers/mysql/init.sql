/*Archivo de Configuraciones que necesitaremos cuando generemos el Docker
Por ejemplo crear la BBDD*/

/* Desactivar el log binario (logs que registran CREATE, UDPATE, ETC) para evitar registrar estas operaciones en los logs de replicacion (Para sincronizar sv secunadrios)
Osea hacemos esto para reducir recursos*/
SET @MYSQLDUMP_TEMP_LOG_BIN = @@SESSION.SQL_LOG_BIN;
SET @@SESSION.SQL_LOG_BIN= 0;

/* Establece el estado de GTIDs purgados, útil en escenarios de replicación. */
SET @@GLOBAL.GTID_PURGED=/*!80000 '+'*/ '';

/*Crea la base de datos go_course_web si no existe.*/
CREATE DATABASE IF NOT EXISTS `go_micro_course`;