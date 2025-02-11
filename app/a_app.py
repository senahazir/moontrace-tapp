import anthropic
import os
from dotenv import load_dotenv
from typing import List, Dict, Optional
from pathlib import Path

# Load environment variables
load_dotenv()

class MoonTrace:
    def __init__(self):
        self.client = anthropic.Anthropic(
            api_key=os.getenv("CLAUDE_API_KEY")
        )
        self.messages: List[Dict[str, str]] = []
        self.system_prompt: str = ""
        
    @staticmethod
    def colorize_text(prompt: str, color_code: str = "33") -> str:
        """Add ANSI color formatting to terminal output."""
        return f"\033[{color_code}m{prompt}\033[0m"

    def read_file_contents(self, file_path: Path) -> Optional[str]:
        """Read file contents with robust error handling."""
        try:
            return file_path.read_text()
        except FileNotFoundError:
            print(f"[Warning] File not found: {file_path}")
            return None
        except Exception as e:
            print(f"[Error] Could not read {file_path}: {e}")
            return None

    def initialize_system_prompt(self, base_dir: Path) -> None:
        """Initialize the system prompt with file contents."""
        required_files = {
            "graph": "graph.txt",
            "analysis": "simulation_log.txt",
            "xml": "Vcounter.xml",
            "vcd": "counter_tb.vcd"
        }
        
        file_contents = {}
        for key, filename in required_files.items():
            content = self.read_file_contents(base_dir / filename)
            if content is None:
                print(f"[Warning] Missing {filename}")
            file_contents[key] = content or ""

        self.system_prompt = f"""
You are MoonTrace, a hardware design engineer with deep knowledge of Verilog, netlists, waveforms, and RTL simulation.

Below is the relevant design data:

===== GRAPH.TXT (Dependency Graph?) =====
{file_contents['graph']}

===== ANALYSIS.TXT (Analysis or notes) =====
{file_contents['analysis']}

===== KEY SIMULATION SNIPPET =====
Time 40: counter.clk changed from 0 to 1.
=> Possibly caused counter.sub_count to change to 0000 at time 40
Time 40: counter.reset changed from 0 to 1.
=> Possibly caused counter.u_counter_monitor.observed_count to change to 1111 at time 40
=> Possibly caused counter.sub_count change to change to 1111 at time 40
=> Possibly caused counter.monitor_flag changed to 0 at time 40

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
10. Unless there is an explicit i -> j in graph.txt, do not assume any i drives any j.
"""

    def process_message(self, user_input: str) -> None:
        """Process user input and get response from Claude."""
        self.messages.append({"role": "user", "content": user_input})

        try:
            response = self.client.messages.create(
                model="claude-3-5-sonnet-20241022",
                max_tokens=200,
                messages=self.messages,
                system=self.system_prompt
            )

            print(self.colorize_text("MoonTrace: ", "36"), end="", flush=True)
            assistant_reply = ""
            print(response.content[0].text)

            # for chunk in response:
            #     content = chunk.choices[0].delta.content
            #     if content:
            #         print(content, end="", flush=True)
            #         assistant_reply += content
            #     elif assistant_reply and not content:
            #         break

            # print()
            self.messages.append({"role": "assistant", "content": assistant_reply})

        except Exception as e:
            print(f"[Error] API call failed: {e}")
            raise

def main():
    moontrace = MoonTrace()
    base_dir = Path("/Users/senagulhazir/Desktop/counter/counter")
    
    print("Welcome to MoonTrace 🌝! Type 'exit' to quit.\n")
    moontrace.initialize_system_prompt(base_dir)

    while True:
        user_input = input(moontrace.colorize_text("You: ", "36"))
        if user_input.strip().lower() == "exit":
            print("Goodbye!")
            break

        try:
            moontrace.process_message(user_input)
        except Exception as e:
            print(f"[Error] Session terminated: {e}")
            break

if __name__ == "__main__":
    main()