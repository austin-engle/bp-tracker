#!/bin/bash

# Default values
DAYS=60
API_URL="http://localhost:32401"

# Function to generate random number in range
random_range() {
    local min=$1
    local max=$2
    echo $((min + RANDOM % (max - min + 1)))
}

# Function to get systolic range based on days ago
get_systolic_range() {
    local days_ago=$1
    local base_range

    if [ $days_ago -gt 45 ]; then
        # Stage 2 hypertension (60-45 days ago)
        base_range="160 180"
    elif [ $days_ago -gt 30 ]; then
        # Stage 1 hypertension (45-30 days ago)
        base_range="145 165"
    elif [ $days_ago -gt 15 ]; then
        # Elevated (30-15 days ago)
        base_range="135 150"
    else
        # Normal (15-1 days ago)
        base_range="120 135"
    fi
    echo $base_range
}

# Function to get diastolic range based on days ago
get_diastolic_range() {
    local days_ago=$1
    local base_range

    if [ $days_ago -gt 45 ]; then
        # Stage 2 hypertension (60-45 days ago)
        base_range="100 110"
    elif [ $days_ago -gt 30 ]; then
        # Stage 1 hypertension (45-30 days ago)
        base_range="90 100"
    elif [ $days_ago -gt 15 ]; then
        # Elevated (30-15 days ago)
        base_range="85 95"
    else
        # Normal (15-1 days ago)
        base_range="75 85"
    fi
    echo $base_range
}

# Function to generate a reading for a specific date
generate_reading() {
    local days_ago=$1

    # Get ranges based on time period
    read -r sys_min sys_max <<< $(get_systolic_range $days_ago)
    read -r dia_min dia_max <<< $(get_diastolic_range $days_ago)

    # Generate base readings with some randomness
    local base_systolic=$(random_range $sys_min $sys_max)
    local base_diastolic=$(random_range $dia_min $dia_max)
    local base_pulse=$(random_range 64 87)

    # Generate timestamp for days ago with random time
    local hour=$(random_range 6 22)
    local minute=$(random_range 0 59)
    local timestamp=$(date -v-${days_ago}d "+%Y-%m-%d ${hour}:%M:00")

    # Generate three readings with slight variations
    local sys1=$base_systolic
    local sys2=$((base_systolic + (RANDOM % 8 - 4)))
    local sys3=$((base_systolic + (RANDOM % 8 - 4)))

    local dia1=$base_diastolic
    local dia2=$((base_diastolic + (RANDOM % 6 - 3)))
    local dia3=$((base_diastolic + (RANDOM % 6 - 3)))

    local pulse1=$base_pulse
    local pulse2=$((base_pulse + (RANDOM % 8 - 4)))
    local pulse3=$((base_pulse + (RANDOM % 8 - 4)))

    response=$(curl -s -X POST -H "Content-Type: application/json" -d "{
        \"timestamp\": \"$timestamp\",
        \"systolic1\": $sys1,
        \"diastolic1\": $dia1,
        \"pulse1\": $pulse1,
        \"systolic2\": $sys2,
        \"diastolic2\": $dia2,
        \"pulse2\": $pulse2,
        \"systolic3\": $sys3,
        \"diastolic3\": $dia3,
        \"pulse3\": $pulse3
    }" http://localhost:32401/submit)

    echo "Added readings for $timestamp: $sys1/$dia1 (p:$pulse1), $sys2/$dia2 (p:$pulse2), $sys3/$dia3 (p:$pulse3)"
}

echo "🌱 Seeding Blood Pressure Data"
echo "============================="

# Parse command line arguments
while [[ "$#" -gt 0 ]]; do
    case $1 in
        --days) DAYS="$2"; shift ;;
        *) echo "Unknown parameter: $1"; exit 1 ;;
    esac
    shift
done

# Generate readings for each day
for (( day=$DAYS; day>=1; day-- )); do
    # Generate 2-4 readings per day
    num_readings=$(random_range 2 4)
    for (( i=1; i<=num_readings; i++ )); do
        generate_reading $day
    done
done

echo
echo "✅ Successfully seeded blood pressure readings showing improvement over time:"
echo "  * Days 60-45: Stage 2 Hypertension (Systolic: 160-180, Diastolic: 100-110)"
echo "  * Days 45-30: Stage 1 Hypertension (Systolic: 145-165, Diastolic: 90-100)"
echo "  * Days 30-15: Elevated           (Systolic: 135-150, Diastolic: 85-95)"
echo "  * Days 15-1:  Normal             (Systolic: 120-135, Diastolic: 75-85)"
