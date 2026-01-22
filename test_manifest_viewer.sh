#!/bin/bash
# Manifest Generation Test Viewer

echo "=== Config Interpolation Tests ==="
go test ./internal/config -v -run TestInterpolateConfig 2>&1 | grep -E "RUN|PASS" | sed 's/=== RUN/  →/; s/--- PASS/  ✓/'

echo ""
echo "=== SteamCMD Command Builder Tests ==="
go test ./internal/steamcmd -v -run TestCommandBuilder 2>&1 | grep -E "RUN|PASS" | sed 's/=== RUN/  →/; s/--- PASS/  ✓/'

echo ""
echo "=== StatefulSet Builder Tests ==="
go test ./internal/resources -v -run TestStatefulSetBuilder_Build 2>&1 | grep -E "RUN|PASS" | head -20 | sed 's/=== RUN/  →/; s/--- PASS/  ✓/'

echo ""
echo "=== Service Builder Tests ==="
go test ./internal/resources -v -run TestServiceBuilder 2>&1 | grep -E "RUN|PASS" | head -15 | sed 's/=== RUN/  →/; s/--- PASS/  ✓/'

echo ""
echo "=== PVC Builder Tests ==="  
go test ./internal/resources -v -run TestPVCBuilder 2>&1 | grep -E "RUN|PASS" | head -15 | sed 's/=== RUN/  →/; s/--- PASS/  ✓/'
