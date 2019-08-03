import React from 'react';
import {BasePageComponent} from "../components/BasePageComponent";

/**
 * @deprecated
 */
export const withBaseComponent = (WrappedComponent: React.ComponentType<any>): React.FC<any> => {
    return () => (
        <BasePageComponent>
            <WrappedComponent />
        </BasePageComponent>
    );
};