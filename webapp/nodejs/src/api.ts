const UserAgent = "isucon9-qualify-webapp";
const IsucariAPIToken = "Bearer 75ugk2m37a750fwir5xr-22l6h4wmue1bwrubzwd0";

export interface ShipmentCreateRequest {
  to_address: string;
  to_name: string;
  from_address: string;
  from_name: string;
}

export interface ShipmentCreateResponse {
  reserve_id: string;
  reserve_time: number;
}

export interface ShipmentRequestRequest {
  reserve_id: string;
}

export interface ShipmentStatusRequest {
  reserve_id: string;
}

export interface ShipmentStatusResponse {
  status: string;
  reserve_time: number;
}

export interface PaymentTokenRequest {
  shop_id: string;
  token: string;
  api_key: string;
  price: number;
}

export interface PaymentTokenResponse {
  status: string;
}

export async function shipmentCreate(url: string, params: ShipmentCreateRequest): Promise<ShipmentCreateResponse> {
  const res = await fetch(url + "/create", {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'User-Agent': UserAgent,
      'Authorization': IsucariAPIToken,
    },
    body: JSON.stringify(params),
  });

  if (!res.ok) {
    throw new Error(`Shipment create failed: ${res.status} ${res.statusText}`);
  }

  return res.json() as Promise<ShipmentCreateResponse>;
}

export async function shipmentRequest(url: string, params: ShipmentRequestRequest): Promise<Buffer> {
  const res = await fetch(url + "/request", {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'User-Agent': UserAgent,
      'Authorization': IsucariAPIToken,
    },
    body: JSON.stringify(params),
  });

  if (!res.ok) {
    throw new Error(`Shipment request failed: ${res.status} ${res.statusText}`);
  }

  const arrayBuffer = await res.arrayBuffer();
  return Buffer.from(arrayBuffer);
}

export async function shipmentStatus(url: string, params: ShipmentStatusRequest): Promise<ShipmentStatusResponse> {
  const res = await fetch(url + "/status", {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'User-Agent': UserAgent,
      'Authorization': IsucariAPIToken,
    },
    body: JSON.stringify(params),
  });

  if (!res.ok) {
    throw new Error(`Shipment status failed: ${res.status} ${res.statusText}`);
  }

  return res.json() as Promise<ShipmentStatusResponse>;
}

export async function paymentToken(url: string, params: PaymentTokenRequest): Promise<PaymentTokenResponse> {
  const res = await fetch(url + "/token", {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'User-Agent': UserAgent,
    },
    body: JSON.stringify(params),
  });

  if (!res.ok) {
    throw new Error(`Payment token failed: ${res.status} ${res.statusText}`);
  }

  return res.json() as Promise<PaymentTokenResponse>;
}
