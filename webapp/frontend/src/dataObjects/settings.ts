import {UserData} from "./user";

export interface Settings {
    csrfToken: string,
    user?: UserData,
}
