import React from 'react';
import BasePageContainer from "../containers/BasePageContainer";
import {ErrorProps, PageComponentWithError} from "../hoc/withBaseComponent";

type Props = {} & ErrorProps

const TransactionPage: React.FC<Props> = () => (
    <BasePageContainer>
        <div>Transaction Page</div>
    </BasePageContainer>
);

export default PageComponentWithError<Props>()(TransactionPage);