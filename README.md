# MoonTrace: Hardware Debugging and Verification Engine 🌝
At its core, MoonTrace is powered by its parsing engine. First half of parsing engine creates a dependency graph for the signals in a verilog/SystemVerilog codebase. Second part of parsing engine traces signals in the simulation output (usually a vcd file that is generated through ModelSim/Verilator .. etc). Using this tracing mechanism, a simulation log is outputted, which user can interact with through a preferred llm to pinpoint errors in design, unexpected behavior, etc. It can also create verification tests as the whole simualtion log is fed into the llm of choice.

NEXT TODO: 
- Add support for creating the test files and running them. 
- Finetune an llm on debugging and verification. 


