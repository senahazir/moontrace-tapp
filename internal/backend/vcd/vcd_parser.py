import re
from collections import defaultdict, deque
import json
import sys
import os

def parse_vcd_to_events(vcd_file_path):
    events = []
    current_time = 0
    id_to_signal = {}
    
    with open(vcd_file_path, 'r') as f:
        for line in f:
            line = line.strip()
            if not line:
                continue
    
            if line.startswith('$var '):
                match = re.match(
                    r'^\$var\s+\S+\s+(\d+)\s+(\S+)\s+(\S+).*\$end',
                    line
                )
                if match:
                    width_str, var_id, var_name = match.groups()
                    id_to_signal[var_id] = var_name
                continue
            
            if line.startswith('#'):
                try:
                    current_time = int(line[1:])
                except ValueError:
                    pass
                continue
            

            if (line[0] in ['0','1']) and len(line) > 1:
                new_val = line[0]
                var_id = line[1:]
                events.append((current_time, var_id, new_val))
                continue
            
            if line.startswith('b'):
                parts = line.split()
                if len(parts) == 2:
                    bin_val = parts[0][1:]  # remember to remove the leading 'b'
                    var_id = parts[1]
                    events.append((current_time, var_id, bin_val))
                continue

    events.sort(key=lambda x: x[0])  # sort by time
    return events, id_to_signal

def label_events_with_names(events, id_to_signal, dependency_graph):

    labeled = []
    for (t, vid, val) in events:
        short_name = id_to_signal.get(vid, f"<unknown:{vid}>")
        
        if short_name in dependency_graph:
            full_name = short_name
        else:
            candidates = [k for k in dependency_graph.keys() if short_name in k]
            values = [k for k in dependency_graph.values()]
            print("values",values)
            if candidates:
                full_name = candidates[0]
            elif values:
                full_name = values[0][0]
            else:
                full_name = short_name  

        labeled.append((t, full_name, val))
    return labeled

def compute_signal_changes(labeled_events):
    changes = []
    current_vals = {}
    for (t, sig, val) in labeled_events:
        old_val = current_vals.get(sig, None)
        if old_val != val:
            changes.append((t, sig, old_val, val))
            current_vals[sig] = val
    return changes


# hopping bfs 

def build_descendants_map(edges):

    from collections import deque, defaultdict

    descendants_of = defaultdict(set)

    for driver in edges.keys():
        visited = set()
        queue = deque([driver])
        while queue:
            current = queue.popleft()
            if current in edges:
                for nxt in edges[current]:
                    if nxt not in visited:
                        visited.add(nxt)
                        queue.append(nxt)
        descendants_of[driver] = visited

    return descendants_of

def analyze_dependencies_possible(changes, edges, time_window=10):

    changes_by_time = defaultdict(list)
    for (t, sig, ov, nv) in changes:
        changes_by_time[t].append((sig, ov, nv))

    descendants_of = build_descendants_map(edges)

    log_messages = []

    all_times = sorted(changes_by_time.keys())

    drivers_by_signal = defaultdict(lambda: defaultdict(set))
    for t in all_times:
        changes = defaultdict(list)
        for (driver_sig, old_val, new_val) in changes_by_time[t]:
            msg = f"Time {t}: {driver_sig} changed from {old_val} to {new_val}."
            log_messages.append(msg)
            
            if driver_sig in descendants_of:
                possible_descendants = descendants_of[driver_sig]
        
                # TODO: remove this timr range or only adjust it to account for 1 clk cycle delays
                for look_time in range(t, t + time_window + 1):
                    if look_time in changes_by_time:
                        for (dsig, d_old, d_new) in changes_by_time[look_time]:
                            print("Dsig", dsig, driver_sig,  possible_descendants)
                            if dsig in possible_descendants:
                                drivers_by_signal[look_time][dsig].add(driver_sig)
                                cmsg = (f"   => {driver_sig} possibly caused {dsig} to change to {d_new} "
                                      f"at time {look_time}")
                                log_messages.append(cmsg)
        if t in drivers_by_signal:
            for signal, drivers in drivers_by_signal[t].items():
                if len(drivers) > 1:
                    multi_driver_msg = f"Time {t}: Signal {signal} has multiple possible drivers: {', '.join(sorted(drivers))}"
                    log_messages.append(multi_driver_msg)

    return log_messages


def main():
    if len(sys.argv) > 1:
        vcd_file = sys.argv[1]
    else:
        vcd_file = "counter_tb.vcd"

    # TODO: update this file path automatically based on users codebase structure 
    dep_json = "dependency_graph.json"
    if not os.path.isfile(dep_json):
        print(f"Error: no '{dep_json}' found.")
        sys.exit(1)

    with open(dep_json, "r") as f:
        dependency_graph = json.load(f)

    if not os.path.isfile(vcd_file):
        print(f"Error: VCD file '{vcd_file}' not found.")
        sys.exit(1)

    events, id_to_signal = parse_vcd_to_events(vcd_file)
    labeled_events = label_events_with_names(events, id_to_signal, dependency_graph)
    changes = compute_signal_changes(labeled_events)
    time_window = 1
    log_messages = analyze_dependencies_possible(changes, dependency_graph, time_window=time_window)

    print("\n=== Multi-Hop Dependency Analysis (Ignoring Intermediate Signals) ===")
    for msg in log_messages:
        print(msg)


if __name__ == "__main__":
    main()
