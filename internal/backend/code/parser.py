import xml.etree.ElementTree as ET
from collections import defaultdict
import json
import sys
import os

def parse_expression(elem, module_hier=""):
    """
    Recursively parse expression nodes (<varref>, etc.) to find driver signals,
    applying 'module_hier' as a prefix if provided.
    """
    sources = set()

    if elem.tag == 'varref':
        sig_name = elem.get('name')
        if sig_name:
            full_name = f"{module_hier}.{sig_name}" if module_hier else sig_name
            sources.add(full_name)
        return sources

    if elem.tag == 'const':
        return sources  # ignore constants

    # Otherwise, recurse on child nodes
    for child in elem:
        sources |= parse_expression(child, module_hier)

    return sources


def parse_contassign(contassign_elem, module_hier=""):
    """
    Parse <contassign>, returning a set of driver signals and a single driven signal,
    each prefixed by 'module_hier' if non-empty.
    """
    children = list(contassign_elem)
    if not children:
        return set(), None

    driven_elem = children[-1]
    driver_elems = children[:-1]

    driver_signals = set()
    for d in driver_elems:
        driver_signals |= parse_expression(d, module_hier)

    driven_signal = None
    if driven_elem.tag == 'varref':
        ds_name = driven_elem.get('name')
        if ds_name:
            driven_signal = f"{module_hier}.{ds_name}" if module_hier else ds_name
    else:
        # Possibly compound expression on the driven side
        driven_signals = parse_expression(driven_elem, module_hier)
        if len(driven_signals) == 1:
            driven_signal = list(driven_signals)[0]
        elif len(driven_signals) > 1:
            driven_signal = ",".join(driven_signals)

    return driver_signals, driven_signal


def parse_assigndly(assigndly_elem, module_hier=""):
    """
    Delayed assignments are basically the same as <contassign> in this structural netlist.
    """
    return parse_contassign(assigndly_elem, module_hier)


def parse_if_block(if_elem, module_hier=""):
    """
    Parse <if> blocks inside <always> statements for conditional assignments.
    """
    driver_signals = set()
    driven_signals = set()

    condition = if_elem.find('varref')
    if condition is not None:
        driver_signals |= parse_expression(condition, module_hier)

    # For each <begin> sub-block, parse <assigndly> or <contassign>
    for begin_block in if_elem.findall('begin'):
        for assigndly_elem in begin_block.findall('.//assigndly'):
            ds, dr = parse_assigndly(assigndly_elem, module_hier)
            driver_signals |= ds
            if dr:
                driven_signals.add(dr)

        for contassign_elem in begin_block.findall('.//contassign'):
            ds, dr = parse_contassign(contassign_elem, module_hier)
            driver_signals |= ds
            if dr:
                driven_signals.add(dr)

    return driver_signals, driven_signals


def parse_instance_ports(instance_elem, parent_hier=""):
    """
    Parse <instance> -> <port> lines, returning a list of
    (submodule_port_full, direction, parent_signal_full).

    'parent_hier' is the *hierarchical path* to this instance's parent.
    We will build 'submodule_prefix' as parent_hier + "." + instance_name
    to properly name signals in the child instance.
    """
    instance_name = instance_elem.get('name')         # e.g. "u_counter_logic"
    def_name = instance_elem.get('defName', None)     # e.g. "counter_logic"
    direction_map = []

    # e.g. "counter.u_counter_logic"
    if parent_hier and instance_name:
        submodule_prefix = f"{parent_hier}.{instance_name}"
    elif instance_name:
        submodule_prefix = instance_name
    else:
        submodule_prefix = parent_hier

    for port in instance_elem.findall('port'):
        port_name = port.get('name')     # e.g. "count"
        direction = port.get('direction','in')  # or "out", "inout"
        varref = port.find('varref')
        if port_name and varref is not None:
            parent_sig_name = varref.get('name')  # e.g. "sub_count"

            # Build child instance port name:
            #   submodule_prefix.port_name
            child_port_full = f"{submodule_prefix}.{port_name}" if submodule_prefix else port_name

            # Parent signal is (parent_hier + "." + parent_sig_name)
            if parent_hier and parent_sig_name:
                parent_sig_full = f"{parent_hier}.{parent_sig_name}"
            else:
                parent_sig_full = parent_sig_name

            direction_map.append((child_port_full, direction, parent_sig_full))
    return direction_map


def parse_verilator_xml_signals(xml_path, top_module_name=None):
    """
    Parse a Verilator XML file into a dictionary: driver_signal -> set of driven_signals,
    using consistent hierarchical naming. We recursively parse:
      - <module name="...">
      - <always>/<if>/<assigndly>/<contassign>
      - <instance name="..." defName="...">

    'top_module_name' can be used if you want to treat one module as top-level.
    Otherwise, we just assume each <module> is a "root" if it's not instantiated by a parent.
    """
    tree = ET.parse(xml_path)
    root = tree.getroot()

    netlist = root.find('netlist')
    if netlist is None:
        raise ValueError("No <netlist> tag found in the XML.")

    edges = defaultdict(set)

    # We'll store 'module_defs' in a dictionary:
    #   module_defs[moduleName] = <module_elem>
    # so we can recursively parse submodules after we parse the top.
    module_defs = {}
    for mod_elem in netlist.findall('module'):
        mname = mod_elem.get('name')
        module_defs[mname] = mod_elem

    # A helper function to parse a *specific module* at a given hierarchical prefix
    def parse_module(mname, hier_prefix):
        """
        Parse the module 'mname' from module_defs, referencing it at hierarchical prefix 'hier_prefix'.
        E.g. if mname="counter_logic" but we are instantiating it as "counter.u_counter_logic".
        We'll parse <always>, <contassign>, <instance> inside it.
        """
        if mname not in module_defs:
            return  # No definition known?

        module_elem = module_defs[mname]

        # 1) Parse <always> blocks
        for always_elem in module_elem.findall('always'):
            # <if> inside <always>
            for if_elem in always_elem.findall('.//if'):
                d_sigs, dr_sigs = parse_if_block(if_elem, hier_prefix)
                for d in d_sigs:
                    for dr in dr_sigs:
                        edges[d].add(dr)

            # <assigndly> inside <always>
            for assigndly_elem in always_elem.findall('.//assigndly'):
                ds, dr = parse_assigndly(assigndly_elem, hier_prefix)
                if dr:
                    for dd in ds:
                        edges[dd].add(dr)

        # 2) Parse top-level <contassign>
        for contassign_elem in module_elem.findall('contassign'):
            ds, dr = parse_contassign(contassign_elem, hier_prefix)
            if dr:
                for dd in ds:
                    edges[dd].add(dr)

        # 3) Parse each <instance> inside this module
        for inst_elem in module_elem.findall('instance'):
            inst_name = inst_elem.get('name')      # e.g. "u_counter_logic"
            defName   = inst_elem.get('defName')   # e.g. "counter_logic"

            # (a) parse instance ports
            ports = parse_instance_ports(inst_elem, hier_prefix)
            for (child_port, direction, parent_sig) in ports:
                if direction == "out":
                    # submodule out -> parent
                    edges[child_port].add(parent_sig)
                else:
                    # parent -> submodule in
                    edges[parent_sig].add(child_port)

            # (b) Recursively parse the submodule definition itself,
            #     giving it a hierarchical prefix = hier_prefix + "." + inst_name
            if defName:
                sub_hier = f"{hier_prefix}.{inst_name}" if hier_prefix else inst_name
                parse_module(defName, sub_hier)

    # Strategy:
    #  - If top_module_name is provided, parse that as the root with the same prefix.
    #  - Otherwise, parse *all* modules as though each is top-level (which may or may not help).
    # Typically, you'd parse just your top-level.

    if top_module_name:
        parse_module(top_module_name, top_module_name)
    else:
        # parse every module as if it were top-level
        for m in module_defs.keys():
            parse_module(m, m)

    return edges


if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Usage: python parse_verilator_xml.py <input_file.xml> [topModuleName]")
        sys.exit(1)

    xml_file = sys.argv[1]
    if not os.path.isfile(xml_file):
        print(f"Error: file '{xml_file}' not found.")
        sys.exit(1)

    maybe_top = None
    if len(sys.argv) > 2:
        maybe_top = sys.argv[2]

    # Parse the Verilator XML with an optional top module name
    signal_deps = parse_verilator_xml_signals(xml_file, top_module_name=maybe_top)

    # Convert sets to lists for JSON serialization
    final_deps = {drv: list(driven) for drv, driven in signal_deps.items()}

    print("Signal-level Dependencies (driver -> driven):")
    for driver, driven_list in final_deps.items():
        print(f"  {driver} -> {driven_list}")

    out_json = "dependency_graph.json"
    with open(out_json, "w") as f:
        json.dump(final_deps, f, indent=2)
    print(f"Dependency graph saved to {out_json}")
