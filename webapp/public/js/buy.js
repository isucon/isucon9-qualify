document.addEventListener('DOMContentLoaded', () => {
    'use strict';

    const shop_id = '11';

    document.getElementById('card_submit').addEventListener('click', (event) => {
        const data = {
            'card_number': document.getElementById('card_number').value,
            'shop_id': shop_id,
        };

        fetch('http://localhost:5555/card', {
            method: 'POST',
            mode: 'cors',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(data),
        }).then((res) => {
            if (!res.ok) {
                throw Error(res.statusText);
            }
            return res.json();
        }).then((json) => {
            const el = document.getElementById('data');
            fetch('/buy', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    csrf_token: el.dataset.csrfToken,
                    item_id: +el.dataset.itemId,
                    token: json.token,
                }),
            });
        }).catch((err) => {
            alert(err);
        });
    });
});
