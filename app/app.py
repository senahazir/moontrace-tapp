import json
import sys
import os
from openai import OpenAI
from dotenv import load_dotenv

# Load API key
load_dotenv()
OpenAI.api_key = os.getenv("OPENAI_API_KEY")
client = OpenAI()


def load_conversation():
    try:
        with open("conversation_history.json", "r") as f:
            return json.load(f)
    except:
        return []

def save_conversation(messages):
    with open("conversation_history.json", "w") as f:
        json.dump(messages, f)

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

def build_system_prompt(base_files, additional_files=None, generate_verification = False, v_filename = None, description = None):
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

    # Add job description for non verification cases  
    if not generate_verification:
        system_prompt += """
                Your job:
                1. Use ONLY the above data for technical details.
                2. Analyze these waveforms and netlists.
                3. If you lack enough data, say so explicitly.
                4. Keep answers concise and accurate.
                5. Keep in mind that you are given a dependency graph and a simulation log that traced events.
                6. You can be given additional files through the process.
                7. Do not use markdown in your answers.
                8. When analyzing dependency, only look at the simulation log.
                9. A signal can become x if there are multiple signals driving it at the same time with different values.
                10. Don't say anything that insinuates that you don't know the cause of a change, but rather give possible reasons. 
                11. if there is undefined/unexpected behavior, explain with the values of the undefined/unexpected variable
                12. When you give code suggestions, explicitly show what you changed. 
                """
    if generate_verification:
        system_prompt += f"""
            You are tasked with generating a comprehensive verification testbench for a hardware design.

            Output Requirements:
            1. Generate a complete, functional C++ testbench for use with Verilator
            2. The testbench should be named "{v_filename if v_filename else 'generated_tb.cpp'}"
            3. Include targeted tests for signals identified in the dependency graph
            4. Create test sequences to expose any X-propagation, timing issues, and edge cases
            5. Test reset behavior thoroughly
            6. Test signals during specific condition changes

            User's Design Description:
            {description if description else "No specific description provided. Create a comprehensive testbench."}

            Your output MUST:
            1. Begin with a C++ testbench header that includes necessary Verilator files
            2. Include multiple test phases (reset testing, normal operation, edge cases)
            3. Have comments explaining test rationale and expectations
            4. Be complete and ready to compile with minimal modifications
            5. Finish with proper cleanup and resource management
            6. Format the entire testbench as a single complete file without markdown formatting

            DO NOT include any explanations or commentary outside the testbench code itself.
            """



    return system_prompt
def process_prompt(user_input, generate_verification, v_filename, description=None, additional_files=None):
    # Define base directory and required files
    global messages
    BASE_DIR = "/Users/senagulhazir/Desktop/demo/"
    base_files = {
        'graph': os.path.join(BASE_DIR, "moontrace/graph.txt"),
        'analysis': "/Users/senagulhazir/Desktop/counter/simulation_log.txt",
        # 'analysis': "/Users/senagulhazir/Desktop/demo/moontrace/simulation_log.txt",
        'xml': os.path.join(BASE_DIR, "counter/Vcounter.xml"),
        'vcd': os.path.join(BASE_DIR, "counter/counter_tb.vcd")
    }

    # Build system prompt with all files
    messages = load_conversation()
    system_prompt = build_system_prompt(base_files, additional_files, generate_verification, v_filename, description)    
    # Message history
    
    if not messages:
        messages = [{"role": "system", "content": system_prompt}]
        messages.append({"role": "user", "content": user_input})
    else:
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

        if generate_verification:
            output_dir = "/Users/senagulhazir/Desktop/demo/counter" 
            if v_filename:
                output_file = os.path.join(output_dir, v_filename)
            else:
                output_file = os.path.join(output_dir, "generated_tb.cpp")

            try:
                with open(output_file, 'w') as f:
                    print(f"\n\nVerification file saved to: {output_file}")
                    f.write(assistant_reply) 
            except Exception as e:
                print(f"\n\nError saving testbench: {e}")
                
        
    except Exception as e:
        print(f"Error: {e}")
    save_conversation(messages)

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Error: Missing prompt argument")
        sys.exit(1)

    user_input = sys.argv[1]
    v_filename = None 
    description = None 
    generate_verification = False
    additional_files = []

    i = 2
    while i < len(sys.argv):
        if sys.argv[i] == "--verification":
            generate_verification = True 
            i += 1 
        elif sys.argv[i] == "--fileName" and i+1 < len(sys.argv):
            v_filename = sys.argv[i + 1]
            i += 2
        elif sys.argv[i] == "--description" and i + 1 < len(sys.argv):
            description =  sys.argv[i + 1]
            i += 2
        else:
            additional_files.append(sys.argv[i]) 
            i += 1 
    
    process_prompt(user_input, generate_verification, v_filename, additional_files)
