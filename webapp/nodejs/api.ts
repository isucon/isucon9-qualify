
import axios, {AxiosResponse} from "axios";

const client = axios.create({ responseType: 'json'});

const UserAgent = "isucon9-qualify-webapp";
const IsucariAPIToken = "Bearer 75ugk2m37a750fwir5xr-22l6h4wmue1bwrubzwd0";

export type ShipmentCreateRequest = {
    to_address: string,
    to_name: string,
    from_address: string,
    from_name: string,
}

export type ShipmentStatusRequest = {
    reserve_id: string,
}

export type ShipmentResponse = {
    status: string,
    reserve_time: number,
}

export async function shipmentCreate(url: string, params: ShipmentCreateRequest): Promise<AxiosResponse> {
    const res = await client.post(url + "/create", params, {
        headers: {
            'User-Agent': UserAgent,
            'Authorization': IsucariAPIToken,
        },
    });
    return res;
}

export async function shipmentStatus(url: string, params: ShipmentStatusRequest): Promise<ShipmentResponse> {
    const res = await client.post(url + "/status", params, {
        headers: {
            'User-Agent': UserAgent,
            'Authorization': IsucariAPIToken,
        },
    });
    if (res.status !== 200) {
        throw res;
    }

    return res.data as ShipmentResponse;
}


