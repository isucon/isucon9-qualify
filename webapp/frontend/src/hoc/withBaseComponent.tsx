import React from 'react';
import {BasePageComponent} from "../components/BasePageComponent";

export const withBaseComponent = (WrappedComponent: React.FC<any>): React.FC<any> => {
    return () => (
        <BasePageComponent>
            <WrappedComponent />
        </BasePageComponent>
    );
};