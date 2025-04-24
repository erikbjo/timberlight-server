import dbf
import datetime

table = dbf.Table("fjordkatalogen_omrade.dbf")
table.open(mode=dbf.READ_WRITE)

for record in dbf.Process(table):
    try:
        float(record.omkrets)
    except (ValueError, TypeError):
        record.omkrets = 0

    try:
        float(record.utmx)
    except (ValueError, TypeError):
        record.utmx = 0

    try:
        float(record.utmy)
    except (ValueError, TypeError):
        record.utmy = 0

table.close()
print("✅ Successfully updated DBF file.")
