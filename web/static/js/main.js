// File: web/static/js/main.js

document.addEventListener('DOMContentLoaded', function() {
    const form = document.getElementById('readingForm');
    const resultDiv = document.getElementById('result');

    form.addEventListener('submit', async function(e) {
        e.preventDefault();

        // Disable submit button while processing
        const submitBtn = form.querySelector('.submit-btn');
        submitBtn.disabled = true;
        submitBtn.textContent = 'Saving...';

        try {
            // Convert form data to JSON
            const formData = new FormData(form);
            const data = {};
            for (let [key, value] of formData.entries()) {
                data[key] = parseInt(value, 10);
            }

            const response = await fetch('/submit', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(data)
            });

            const result = await response.json();

            if (response.ok) {
                // Show success message and classification
                displayResult(result, false);
                updateStatsDisplay(result.stats);
                form.reset();

                // Refresh the stats section
                const statsSection = document.querySelector('.stats-grid');
                if (statsSection && result.stats) {
                    updateStatsDisplay(result.stats);
                }
            } else {
                // Show error message
                displayResult(result, true);
            }
        } catch (error) {
            console.error('Error:', error);
            displayResult({
                error: 'Error submitting readings. Please try again.'
            }, true);
        } finally {
            // Re-enable submit button
            submitBtn.disabled = false;
            submitBtn.textContent = 'Save Readings';
        }
    });

    function updateStatsDisplay(stats) {
        // Create or update stats section
        let statsGrid = document.querySelector('.stats-grid');
        if (!statsGrid) {
            // Create stats section if it doesn't exist
            const statsSection = document.querySelector('.stats-section');
            statsGrid = document.createElement('div');
            statsGrid.className = 'stats-grid';
            statsSection.prepend(statsGrid);
        }

        // Create stats cards if they don't exist
        if (stats.seven_day_avg && !document.querySelector('.stat-card')) {
            const averages = [
                { title: '7-Day Average', data: stats.seven_day_avg },
                { title: '30-Day Average', data: stats.thirty_day_avg },
                { title: 'All-Time Average', data: stats.all_time_avg }
            ];

            averages.forEach(avg => {
                const card = document.createElement('div');
                card.className = 'stat-card';
                card.innerHTML = `
                    <h3>${avg.title}</h3>
                    <p>${avg.data.systolic}/${avg.data.diastolic} mmHg</p>
                    <p>Pulse: ${avg.data.pulse} bpm</p>
                `;
                statsGrid.appendChild(card);
            });
        } else {
            // Update existing cards
            updateAverageCard('7-Day Average', stats.seven_day_avg);
            updateAverageCard('30-Day Average', stats.thirty_day_avg);
            updateAverageCard('All-Time Average', stats.all_time_avg);
        }
    }

    function updateAverageCard(title, data) {
        if (!data) return;
        const cards = document.querySelectorAll('.stat-card');
        for (const card of cards) {
            if (card.querySelector('h3').textContent === title) {
                const paragraphs = card.querySelectorAll('p');
                if (paragraphs.length >= 2) {
                    paragraphs[0].textContent = `${data.systolic}/${data.diastolic} mmHg`;
                    paragraphs[1].textContent = `Pulse: ${data.pulse} bpm`;
                }
                break;
            }
        }
    }

    function displayResult(result, isError) {
        resultDiv.classList.remove('hidden');
        const classificationEl = resultDiv.querySelector('.classification');
        const recommendationEl = resultDiv.querySelector('.recommendation');

        if (isError) {
            classificationEl.textContent = `Error: ${result.error}`;
            classificationEl.className = 'classification crisis';
            recommendationEl.textContent = '';
        } else {
            classificationEl.textContent = `Classification: ${result.classification.Name}`;
            classificationEl.className = `classification ${result.classification.Name.toLowerCase().replace(' ', '')}`;
            recommendationEl.textContent = result.recommendation;
        }
    }
});
