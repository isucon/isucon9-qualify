import React from 'react';
import BasePageContainer from "../containers/BasePageContainer";

/**
 * @deprecated
 */
export const withBaseComponent = (WrappedComponent: React.ComponentType<any>): React.FC<any> => {
    return () => (
        <BasePageContainer>
            <WrappedComponent />
        </BasePageContainer>
    );
};