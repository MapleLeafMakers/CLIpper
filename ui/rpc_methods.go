package ui

var MoonrakerRPCMethods = []string{
	"access.delete_user",
	"access.get_api_key",
	"access.get_user",
	"access.info",
	"access.login",
	"access.logout",
	"access.oneshot_token",
	"access.post_api_key",
	"access.post_user",
	"access.refresh_jwt",
	"access.users.list",
	"access.users.password",
	"connection.register_remote_method",
	"connection.send_event",
	"debug.database.delete_item",
	"debug.database.get_item",
	"debug.database.list",
	"debug.database.post_item",
	"debug.notifiers.test",
	"machine.device_power.devices",
	"machine.device_power.get_device",
	"machine.device_power.off",
	"machine.device_power.on",
	"machine.device_power.post_device",
	"machine.device_power.status",
	"machine.proc_stats",
	"machine.reboot\t",
	"machine.services.restart",
	"machine.services.start",
	"machine.services.stop",
	"machine.shutdown",
	"machine.sudo.info",
	"machine.sudo.password",
	"machine.system_info",
	"machine.update.client",
	"machine.update.full",
	"machine.update.klipper",
	"machine.update.moonraker",
	"machine.update.recover",
	"machine.update.refresh",
	"machine.update.rollback",
	"machine.update.status",
	"machine.update.system",
	"machine.wled.get_strip",
	"machine.wled.off",
	"machine.wled.on",
	"machine.wled.post_strip",
	"machine.wled.status",
	"machine.wled.strips",
	"machine.wled.toggle",
	"printer.emergency_stop",
	"printer.firmware_restart",
	"printer.gcode.help",
	"printer.gcode.script",
	"printer.info",
	"printer.objects.list",
	"printer.objects.query",
	"printer.objects.subscribe",
	"printer.print.cancel",
	"printer.print.pause",
	"printer.print.resume",
	"printer.print.start",
	"printer.query_endstops.status",
	"printer.restart",
	"server.announcements.delete_feed",
	"server.announcements.dismiss",
	"server.announcements.feeds",
	"server.announcements.list",
	"server.announcements.post_feed",
	"server.announcements.update",
	"server.config",
	"server.connection.identify",
	"server.database.delete_item",
	"server.database.get_item",
	"server.database.list",
	"server.database.post_item",
	"server.extensions.list",
	"server.extensions.request",
	"server.files.copy",
	"server.files.delete_directory",
	"server.files.delete_file",
	"server.files.get_directory",
	"server.files.list",
	"server.files.metadata",
	"server.files.metascan",
	"server.files.move",
	"server.files.post_directory",
	"server.files.roots",
	"server.files.thumbnails",
	"server.files.zip",
	"server.gcode_store",
	"server.hisory.delete_job",
	"server.history.get_job",
	"server.history.list",
	"server.history.reset_totals",
	"server.history.totals",
	"server.info",
	"server.job_queue.delete_job",
	"server.job_queue.jump",
	"server.job_queue.pause",
	"server.job_queue.post_job",
	"server.job_queue.start",
	"server.job_queue.status",
	"server.logs.rollover",
	"server.mqtt.publish",
	"server.mqtt.subscribe",
	"server.notifiers.list",
	"server.restart",
	"server.sensors.info",
	"server.sensors.list",
	"server.sensors.measurements",
	"server.spoolman.get_spool_id",
	"server.spoolman.post_spool_id",
	"server.spoolman.proxy",
	"server.temperature_store",
	"server.webcams.delete_item",
	"server.webcams.get_item",
	"server.webcams.list",
	"server.webcams.post_item",
	"server.webcams.test",
	"server.websocket.id",
}