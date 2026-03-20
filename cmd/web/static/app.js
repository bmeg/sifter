document.addEventListener('DOMContentLoaded', () => {
    const listEl = document.getElementById('playbook-list');
    const contentEl = document.getElementById('playbook-content');

    // Fetch list of playbooks
    fetch('/api/playbooks')
        .then(res => res.json())
        .then(playbooks => {
            playbooks.forEach(name => {
                const li = document.createElement('li');
                li.textContent = name;
                li.addEventListener('click', () => loadPlaybook(name, li));
                listEl.appendChild(li);
            });
        })
        .catch(err => console.error('Failed to load playbook list', err));

    function loadPlaybook(name, element) {
        // Highlight selected
        Array.from(listEl.children).forEach(child => child.classList.remove('selected'));
        element.classList.add('selected');
        fetch(`/api/playbook?name=${encodeURIComponent(name)}`)
            .then(res => {
                if (!res.ok) throw new Error('Playbook not found');
                return res.text();
            })
            .then(text => {
                contentEl.textContent = text;
                hljs.highlightElement(contentEl);
            })
            .catch(err => console.error('Failed to load playbook', err));
    }
});
