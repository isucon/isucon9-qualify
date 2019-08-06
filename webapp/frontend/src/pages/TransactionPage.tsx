import React from 'react';
import {BasePageComponent} from "../components/BasePageComponent";
import {ErrorProps, PageComponentWithError} from "../hoc/withBaseComponent";
import LoadingComponent from "../components/LoadingComponent";

type Props = {
    loading: boolean,
} & ErrorProps

const TransactionPage: React.FC<Props> = ({ loading }) => (
    <BasePageComponent>
        {
            loading ? (
                <LoadingComponent/>
            ) : (
                <div>Transaction Page</div>
            )
        }
    </BasePageComponent>
);

export default PageComponentWithError<Props>()(TransactionPage);