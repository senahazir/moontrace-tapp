from openai import OpenAI
import os
from dotenv import load_dotenv


load_dotenv()
OpenAI.api_key = os.getenv("OPENAI_API_KEY")


client = OpenAI()

def colorize_text(prompt, color_code = "33"):
    return f"\033[{color_code}m{prompt}\033[0m"




## # ===== XML FILE (Netlist) =====
# {xml_content}

# ===== VCD FILE (Waveforms) =====
# {vcd_content}

def main():

    


    print("Welcome to MoonTrace 🌝! Type 'exit' to quit.\n")

    def read_file_contents(file_path):

        try:
            with open(file_path, 'r') as f:
                return f.read()
        except FileNotFoundError:
            print(f"[Warning] File not found: {file_path}")
            return ""
        except Exception as e:
            print(f"[Error] Could not read {file_path}: {e}")
            return ""


    BASE_DIR = "/Users/senagulhazir/Desktop/counter/counter"

    # with open(os.path.join(BASE_DIR, "simulation_log.txt"), 'r') as file:
    #     lines = file.readlines()
    #     print("HELLO", lines[835])  

    xml_content = read_file_contents(os.path.join(BASE_DIR, "Vcounter.xml"))     
    vcd_content = read_file_contents(os.path.join(BASE_DIR, "counter_tb.vcd"))
    graph_content = read_file_contents(os.path.join(BASE_DIR, "graph.txt"))
    analysis_content = read_file_contents(os.path.join(BASE_DIR, "simulation_log.txt"))


    system_prompt = f"""
You are a hardware design engineer with deep knowledge of Verilog, netlists, waveforms, and RTL simulation.

Below is the relevant design data:



===== GRAPH.TXT (Dependency Graph?) =====
{graph_content}

===== ANALYSIS.TXT (Analysis or notes) =====
{analysis_content}

===== KEY SIMULATION SNIPPET ===== Time 40: counter.clk changed from 0 to 1. => Possibly caused counter.sub_count to change to 0000 at time 40 Time 40: counter.reset changed from 0 to 1. => Possibly caused counter.u_counter_monitor.observed_count to change to 1111 at time 40 => Possibly caused counter.sub_count change to change to 1111 at time 40 => Possibly caused counter.monitor_flag changed to 0 at time 40 Time 40: counter.u_counter_monitor.observed_count changed from 0000 to 1111. Time 40: counter.sub_count changed from 0000 to xxxx. => Possibly caused counter.u_counter_monitor.observed_count changed to 1111 at time 40 => Possibly caused counter.monitor_flag changed to 0 at time 40 Time 40: counter.monitor_flag changed from 1 to 0.



Your job:
1. Use ONLY the above data for technical details.
2. Analyze these waveforms and netlists.
3. If you lack enough data, say so explicitly.
4. Keep answers concise and accurate.
5. Keep in mind that you are given a dependency graph and a simulation log that traced events, so based on the questions you receive, please carefully review those two files 
6. You can be given additional files through the process, make sure to analyze them carefully as well 
7. Do not use markdown in your answers
8. when analyzing dependency, only look at the simulation log, ie a signal i only potentially affects a signal j at time t only if there is a line that states i potentially affected j 
9. a signal can become x if there are multiple signals driving it at the same time with different values. 
10. Unless there is an explicit i -> j in graph.txt, do not assume any i drives any j. J only depends on i if there is i -> j in
"""


    messages = [
        {"role": "system", "content": system_prompt}
    ]

    while True:
        user_input = input(colorize_text("You: ", "36"))
        if user_input.strip().lower() == "exit":
            print("Goodbye!")
            break
        messages.append({"role": "user", "content": user_input})

        try:
            response = client.chat.completions.create(
            model="gpt-4o-mini",  
            messages=messages, 
            stream = True,
        )
            #assistant_reply = response.choices[0].message.content
            assistant_reply = ""
            print(colorize_text("MoonTrace: ", "36"), end="", flush=True)
            content = ""
            for chunk in response:
                content = chunk.choices[0].delta.content
                if content:
                    print(content, end="", flush=True)
                    assistant_reply += content
                elif assistant_reply and not content:
                    break

            print()

            messages.append({"role": "assistant", "content": assistant_reply})

        except Exception as e:
            print(f"[Error] OpenAI API call failed: {e}")
            break

if __name__ == "__main__":
    main()