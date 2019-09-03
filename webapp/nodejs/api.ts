
import axios, {AxiosResponse} from "axios";

const client = axios.create({ responseType: 'json'});

const UserAgent = "isucon9-qualify-webapp";
const IsucariAPIToken = "Bearer 75ugk2m37a750fwir5xr-22l6h4wmue1bwrubzwd0";

type ShipmentCreateRequest = {
    to_address: string,
    to_name: string,
    from_address: string,
    from_name: string,
}

type ShipmentCreateResponse = {
    reserve_id: string,
    reserve_time: number,
}

type ShipmentRequestRequest = {
    reserve_id: string,
}

type ShipmentStatusRequest = {
    reserve_id: string,
}

type ShipmentStatusResponse = {
    status: string,
    reserve_time: number,
}

type PaymentTokenRequest = {
    shop_id: string,
    token: string,
    api_key: string,
    price: number,
}

type PaymentTokenResponse = {
    status: string,
}

export async function shipmentCreate(url: string, params: ShipmentCreateRequest): Promise<ShipmentCreateResponse> {
    const res = await client.post(url + "/create", params, {
        headers: {
            'User-Agent': UserAgent,
            'Authorization': IsucariAPIToken,
        },
    });
    if (res.status !== 200) {
        throw res;
    }

    return res.data as ShipmentCreateResponse;
}

export async function shipmentRequest(url: string, params: ShipmentRequestRequest): Promise<ArrayBuffer> {
    const res = await client.post(url + "/request", params, {
        responseType: 'arraybuffer',
        headers: {
            'User-Agent': UserAgent,
            'Authorization': IsucariAPIToken,
        },
    });
    if (res.status !== 200) {
        throw res;
    }

    return res.data;
}

export async function shipmentStatus(url: string, params: ShipmentStatusRequest): Promise<ShipmentStatusResponse> {
    const res = await client.post(url + "/status", params, {
        headers: {
            'User-Agent': UserAgent,
            'Authorization': IsucariAPIToken,
        },
    });
    if (res.status !== 200) {
        throw res;
    }

    return res.data as ShipmentStatusResponse;
}

export async function paymentToken(url: string, params: PaymentTokenRequest): Promise<PaymentTokenResponse> {
    const res = await client.post(url + "/token", params, {
        headers: {
            'User-Agent': UserAgent,
        },
    });
    if (res.status !== 200) {
        throw res;
    }

    return res.data as PaymentTokenResponse;
}


