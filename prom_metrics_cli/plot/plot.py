from matplotlib import container
import matplotlib.pyplot as plt
import json 
import os
import argparse

parser = argparse.ArgumentParser(description='Parse matplotlib labels.')
parser.add_argument('-x', action='store', type=str,
                    help='xlabel')
parser.add_argument('-y', action='store', type=str,
                    help='ylabel')
parser.add_argument('-l', '--legend-list', nargs='+', default=[], help="Manually import a list of legends for matplotlib")
parser.add_argument('-f', action='store', help="Json field from which to extract legends")

args = parser.parse_args()

dirname, _ = os.path.split(os.path.abspath(__file__))

data = input()

if " " in data:
    data = data.split(" ")[2]
try:
    data = json.loads(data)
except:
    print("Input data not json serializable")

legend = []    
for result in data["data"]["result"]:

    base = float(result["values"][0][0])
    time = []
    position = []
    for v in result["values"]:
        time.append(float(v[0]) - base)
        position.append(float(v[1]))
            
    plt.plot(time, position)
    plt.xlabel(args.x)
    plt.ylabel(args.y)
    if args.f is not None:
        legend.append(result["metric"][args.f])

    if not os.path.exists(dirname + "/figures"):
        os.makedirs(dirname + "/figures")
    
    max_id = 0
    for filename in os.listdir(dirname + "/figures"):
        try:
            tmp = int(filename.split(".")[0])
        except:
            continue

        if tmp > max_id:
            max_id = tmp

if len(args.legend_list) > 0:
    legend = []
    for l in args.legend_list:
        legend.append(l)
    plt.legend(legend)
elif args.f is not None:
    plt.legend(legend)

plt.savefig(dirname + '/figures/' + str(max_id + 1) + '.png')