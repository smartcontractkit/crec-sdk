import sys
import re
from collections import defaultdict

def main():
    if len(sys.argv) < 2:
        return

    filepath = sys.argv[1]
    with open(filepath, 'r') as f:
        lines = f.readlines()

    # Parse commits
    commits = []
    other_lines = []
    
    # regex for: pick <hash> <message>
    pattern = re.compile(r'^(\w+)\s+([0-9a-f]+)\s+(.+)$')
    
    for line in lines:
        if line.strip().startswith('#') or not line.strip():
            other_lines.append(line)
            continue
            
        match = pattern.match(line)
        if match:
            cmd, hsh, msg = match.groups()
            commits.append((cmd, hsh, msg.strip()))
        else:
            other_lines.append(line)

    # Group by message preserving order of first appearance
    message_groups = defaultdict(list)
    ordered_messages = []
    
    for cmd, hsh, msg in commits:
        if msg not in message_groups:
            ordered_messages.append(msg)
        message_groups[msg].append((cmd, hsh))

    # Reconstruct the file content
    with open(filepath, 'w') as f:
        for msg in ordered_messages:
            group = message_groups[msg]
            # First one is picked
            first_cmd, first_hsh = group[0]
            f.write(f"pick {first_hsh} {msg}\n")
            
            # Rest are fixed up
            for cmd, hsh in group[1:]:
                f.write(f"fixup {hsh} {msg}\n")
                
        # Write back comments and empty lines
        for line in other_lines:
            f.write(line)

if __name__ == "__main__":
    main()