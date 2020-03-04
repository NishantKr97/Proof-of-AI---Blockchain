import csv

with open('data.txt', 'r') as in_file:
    stripped = (line.strip() for line in in_file)
    lines = (line.split(",") for line in stripped if line)
    with open('data.csv', 'w') as out_file:
        writer = csv.writer(out_file)
        writer.writerow(('UserId', 'No of Blocks Mined', 'Wrong Transactions', 'Time', 'Age'))
        writer.writerows(lines)