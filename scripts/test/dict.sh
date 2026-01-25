#!/bin/bash
# Dictionary Tests - Combined Summary
# Thresholds: VNI 100%, Telex 100%, Telex+AR 100%, EN 97%

cd "$(dirname "$0")/../../core"

# Run VNI test
VNI_OUTPUT=$(cargo test --test vietnamese_dict_test vietnamese_dict_vni -- --exact --nocapture 2>&1 || true)
VNI_TOTAL=$(echo "$VNI_OUTPUT" | grep "Total words" | grep -oE '[0-9]+' | tail -1)
VNI_PASSED=$(echo "$VNI_OUTPUT" | grep "Passed" | grep -oE '[0-9]+' | tail -1)
VNI_FAILED=$(echo "$VNI_OUTPUT" | grep "Failed" | grep -oE '[0-9]+' | tail -1)
VNI_RATE=$(echo "scale=2; $VNI_PASSED * 100 / $VNI_TOTAL" | bc)

# Run Telex test
TELEX_OUTPUT=$(cargo test --test vietnamese_dict_test vietnamese_dict_telex -- --exact --nocapture 2>&1 || true)
TELEX_TOTAL=$(echo "$TELEX_OUTPUT" | grep "Total words" | grep -oE '[0-9]+' | tail -1)
TELEX_PASSED=$(echo "$TELEX_OUTPUT" | grep "Passed" | grep -oE '[0-9]+' | tail -1)
TELEX_FAILED=$(echo "$TELEX_OUTPUT" | grep "Failed" | grep -oE '[0-9]+' | tail -1)
TELEX_RATE=$(echo "scale=2; $TELEX_PASSED * 100 / $TELEX_TOTAL" | bc)

# Run Telex with auto-restore test
TELEX_AR_OUTPUT=$(cargo test --test vietnamese_dict_test vietnamese_dict_telex_auto_restore -- --exact --nocapture 2>&1 || true)
TELEX_AR_TOTAL=$(echo "$TELEX_AR_OUTPUT" | grep "Total words" | grep -oE '[0-9]+' | tail -1)
TELEX_AR_PASSED=$(echo "$TELEX_AR_OUTPUT" | grep "Passed" | grep -oE '[0-9]+' | tail -1)
TELEX_AR_FAILED=$(echo "$TELEX_AR_OUTPUT" | grep "Failed" | grep -oE '[0-9]+' | tail -1)
TELEX_AR_RATE=$(echo "scale=2; $TELEX_AR_PASSED * 100 / $TELEX_AR_TOTAL" | bc)

# Run English test
EN_OUTPUT=$(cargo test --test english_100k_test english_100k_failures -- --nocapture 2>&1 || true)
EN_TOTAL=$(echo "$EN_OUTPUT" | grep "Total words" | grep -oE '[0-9]+' | tail -1)
EN_PASSED=$(echo "$EN_OUTPUT" | grep "Passed" | grep -oE '[0-9]+' | tail -1)
EN_FAILED=$(echo "$EN_OUTPUT" | grep "Failed" | grep -oE '[0-9]+' | tail -1)
EN_RATE=$(echo "scale=2; $EN_PASSED * 100 / $EN_TOTAL" | bc)

# Extract failure causes
EN_TONE=$(echo "$EN_OUTPUT" | grep "Tone (s/f" | grep -oE '[0-9]+' | head -1)
EN_VOWEL=$(echo "$EN_OUTPUT" | grep "Vowel (aa" | grep -oE '[0-9]+' | head -1)
EN_BOTH=$(echo "$EN_OUTPUT" | grep "│ Both" | grep -oE '[0-9]+' | head -1)
EN_UNKNOWN=$(echo "$EN_OUTPUT" | grep "│ Unknown" | grep -oE '[0-9]+' | head -1)

# Calculate cause percentages
TONE_PCT=$(echo "scale=1; $EN_TONE * 100 / $EN_FAILED" | bc)
VOWEL_PCT=$(echo "scale=1; $EN_VOWEL * 100 / $EN_FAILED" | bc)
BOTH_PCT=$(echo "scale=1; $EN_BOTH * 100 / $EN_FAILED" | bc)
UNKNOWN_PCT=$(echo "scale=1; $EN_UNKNOWN * 100 / $EN_FAILED" | bc)

# Calculate Vietnamese aggregate (worst case)
VN_TOTAL=$VNI_TOTAL
VN_FAILED=$((VNI_FAILED + TELEX_FAILED + TELEX_AR_FAILED))
VN_PASSED=$((VN_TOTAL * 3 - VN_FAILED))
# Use worst rate among the 3 tests
VN_RATE=$(echo "scale=2; if ($VNI_RATE < $TELEX_RATE) { if ($VNI_RATE < $TELEX_AR_RATE) $VNI_RATE else $TELEX_AR_RATE } else { if ($TELEX_RATE < $TELEX_AR_RATE) $TELEX_RATE else $TELEX_AR_RATE }" | bc)

# Print combined summary table
echo ""
echo "┌──────────────────────────────────────────────────────────────────┐"
echo "│                    DICTIONARY TEST RESULTS                       │"
echo "├──────────────┬──────────┬──────────┬──────────┬─────────┬────────┤"
echo "│ Dictionary   │ Total    │ Passed   │ Failed   │ Rate    │ Target │"
echo "├──────────────┼──────────┼──────────┼──────────┼─────────┼────────┤"
printf "│ Vietnamese   │ %8s │ %8s │ %8s │ %6s%% │   100%% │\n" "$VN_TOTAL" "$VN_PASSED" "$VN_FAILED" "$VN_RATE"
printf "│  - VNI       │          │          │ %8s │ %6s%% │        │\n" "$VNI_FAILED" "$VNI_RATE"
printf "│  - Telex     │          │          │ %8s │ %6s%% │        │\n" "$TELEX_FAILED" "$TELEX_RATE"
printf "│  - Telex+AR  │          │          │ %8s │ %6s%% │        │\n" "$TELEX_AR_FAILED" "$TELEX_AR_RATE"
echo "├──────────────┼──────────┼──────────┼──────────┼─────────┼────────┤"
printf "│ English      │ %8s │ %8s │ %8s │ %6s%% │    97%% │\n" "$EN_TOTAL" "$EN_PASSED" "$EN_FAILED" "$EN_RATE"
printf "│  - Tone      │          │          │ %8s │ %6s%% │        │\n" "$EN_TONE" "$TONE_PCT"
printf "│  - Vowel     │          │          │ %8s │ %6s%% │        │\n" "$EN_VOWEL" "$VOWEL_PCT"
printf "│  - Both      │          │          │ %8s │ %6s%% │        │\n" "$EN_BOTH" "$BOTH_PCT"
printf "│  - Unknown   │          │          │ %8s │ %6s%% │        │\n" "$EN_UNKNOWN" "$UNKNOWN_PCT"
echo "└──────────────┴──────────┴──────────┴──────────┴─────────┴────────┘"

# Check thresholds
VNI_OK=$(echo "$VNI_RATE >= 100" | bc)
TELEX_OK=$(echo "$TELEX_RATE >= 100" | bc)
TELEX_AR_OK=$(echo "$TELEX_AR_RATE >= 100" | bc)
EN_OK=$(echo "$EN_RATE >= 97" | bc)

if [ "$VNI_OK" -eq 1 ] && [ "$TELEX_OK" -eq 1 ] && [ "$TELEX_AR_OK" -eq 1 ] && [ "$EN_OK" -eq 1 ]; then
    echo ""
    echo "✅ All dictionary tests passed"
    exit 0
else
    echo ""
    [ "$VNI_OK" -eq 0 ] && echo "❌ VNI: $VNI_RATE% < 100%"
    [ "$TELEX_OK" -eq 0 ] && echo "❌ Telex: $TELEX_RATE% < 100%"
    [ "$TELEX_AR_OK" -eq 0 ] && echo "❌ Telex+AutoRestore: $TELEX_AR_RATE% < 100%"
    [ "$EN_OK" -eq 0 ] && echo "❌ English: $EN_RATE% < 97%"
    exit 1
fi
