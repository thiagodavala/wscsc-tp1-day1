import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
    vus: 10, // Número de usuários virtuais (ajuste conforme necessário)
    duration: '4h', // Duração do teste
};

function getRandomInt(min, max) {
    min = Math.ceil(min);
    max = Math.floor(max);
    return Math.floor(Math.random() * (max - min + 1)) + min;
}

export default function () {
    const value = getRandomInt(100000, 200000);
    const url = `http://localhost:8081/calc?input=${value}`; // Substitua com a URL correta

    const response = http.get(url);

    check(response, {
        'status is 200': (r) => r.status === 200,
        'transaction time ok': (r) => r.timings.duration < 500,
    });

    sleep(1); // Tempo de espera entre as requisições (ajuste conforme necessário)
}
