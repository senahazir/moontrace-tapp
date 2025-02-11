import sys
import json
import requests
import os
from dotenv import load_dotenv

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

def colorize_text(prompt, color_code = "33"):
    return f"\033[{color_code}m{prompt}\033[0m"

def process_prompt(user_input, messages):
    payload = {
        "model": "deepseek-coder",
        "prompt": f"Previous conversation:\n{messages}\n\nUser: {user_input}\nAssistant:"
    }
    
    print(colorize_text("MoonTrace: ", "36"), end="", flush=True)
    
    response = requests.post('http://localhost:11434/api/generate', 
                           json=payload,
                           stream=True)
    
    if response.status_code == 200:
        assistant_reply = ""
        for line in response.iter_lines():
            if line:
                json_response = json.loads(line)
                if not json_response.get('done', False):
                    chunk = json_response.get('response', '')
                    print(chunk, end='', flush=True)
                    assistant_reply += chunk
        print()
        return assistant_reply
    else:
        print(f"Error: {response.status_code}")
        return None

def main():
    print("Welcome to MoonTrace 🌝! Type 'exit' to quit.\n")
    
    BASE_DIR = "/Users/senagulhazir/Desktop/counter/counter"
    xml_content = read_file_contents(os.path.join(BASE_DIR, "Vcounter.xml"))
    vcd_content = read_file_contents(os.path.join(BASE_DIR, "counter_tb.vcd"))
    graph_content = read_file_contents(os.path.join(BASE_DIR, "graph.txt"))
    analysis_content = read_file_contents(os.path.join(BASE_DIR, "simulation_log.txt"))
    
    system_prompt = f"""You are a hardware design engineer with deep knowledge of Verilog, netlists, waveforms, and RTL simulation.
===== GRAPH.TXT =====
{graph_content}
===== ANALYSIS.TXT =====
{analysis_content}
Your job: Analyze these files and answer questions concisely and accurately."""
    
    conversation = [system_prompt]
    
    while True:
        user_input = input(colorize_text("You: ", "36"))
        if user_input.strip().lower() == "exit":
            print("Goodbye!")
            break
            
        conversation.append(f"User: {user_input}")
        reply = process_prompt(user_input, "\n".join(conversation))
        if reply:
            conversation.append(f"Assistant: {reply}")

if __name__ == "__main__":
    main()