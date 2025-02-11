import sys
import os
from openai import OpenAI
from dotenv import load_dotenv

# Load API key
load_dotenv()
OpenAI.api_key = os.getenv("OPENAI_API_KEY")
client = OpenAI()

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

def build_system_prompt(base_files, additional_files=None):
    # Read base files
    base_content = {}
    for name, path in base_files.items():
        content = read_file_contents(path)
        base_content[name] = content

    # Start with base system prompt
    system_prompt = f"""
You are a hardware design engineer with deep knowledge of Verilog, netlists, waveforms, and RTL simulation.

Relevant data:

===== GRAPH.TXT =====
{base_content['graph']}

===== ANALYSIS.TXT =====
{base_content['analysis']}
"""

    # Add any additional files that were selected
    if additional_files:
        for file_path in additional_files:
            content = read_file_contents(file_path)
            if content:
                file_name = os.path.basename(file_path).upper()
                system_prompt += f"\n===== {file_name} =====\n{content}\n"

    # Add job description
    system_prompt += """
Your job:
1. Use ONLY the above data for technical details.
2. Analyze these waveforms and netlists.
3. If you lack enough data, say so explicitly.
4. Keep answers concise and accurate.
5. Keep in mind that you are given a dependency graph and a simulation log that traced events.
6. You can be given additional files through the process, make sure to analyze them carefully as well.
7. Do not use markdown in your answers.
"""
    return system_prompt

def process_prompt(user_input, additional_files=None):
    # Define base directory and required files
    BASE_DIR = "/Users/senagulhazir/Desktop/counter/counter"
    base_files = {
        'graph': os.path.join(BASE_DIR, "graph.txt"),
        'analysis': os.path.join(BASE_DIR, "simulation_log.txt"),
        'xml': os.path.join(BASE_DIR, "Vcounter.xml"),
        'vcd': os.path.join(BASE_DIR, "counter_tb.vcd")
    }

    # Build system prompt with all files
    system_prompt = build_system_prompt(base_files, additional_files)
    
    # Message history
    messages = [{"role": "system", "content": system_prompt}]
    messages.append({"role": "user", "content": user_input})

    try:
        response = client.chat.completions.create(
            model="gpt-4o-mini",
            messages=messages,
            stream=True
        )
        
        assistant_reply = ""
        content = ""
        for chunk in response:
            content = chunk.choices[0].delta.content
            if content:
                print(content, end="", flush=True)
                assistant_reply += content
            elif assistant_reply and not content:
                break
                
        messages.append({"role": "assistant", "content": assistant_reply})

        
    except Exception as e:
        print(f"Error: {e}")

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Error: Missing prompt argument")
        sys.exit(1)

    user_input = sys.argv[1]
    # Any additional arguments are file paths
    additional_files = sys.argv[2:] if len(sys.argv) > 2 else None
    
    process_prompt(user_input, additional_files)