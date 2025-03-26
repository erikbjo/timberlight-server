import dbf
import datetime

table = dbf.Table("/path/to/LosmasseGrense_yyyymmdd.dbf")
table.open(mode=dbf.READ_WRITE)

for record in dbf.Process(table):
    if record.OPPDATERIN in [None, "    "]:
        record.OPPDATERIN = datetime.date(2025, 1, 1)
    if record.FORSTEDIGI in [None, "    "]:
        record.FORSTEDIGI = datetime.date(2025, 1, 1)

table.close()
print("✅ Successfully updated OPPDATERIN and FORSTEDIGI in DBF file.")
