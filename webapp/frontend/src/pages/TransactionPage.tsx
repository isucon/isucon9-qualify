import React, { ReactElement } from 'react';
import BasePageContainer from '../containers/BasePageContainer';
import { ErrorProps, PageComponentWithError } from '../hoc/withBaseComponent';
import { RouteComponentProps } from 'react-router';
import { ItemData } from '../dataObjects/item';
import LoadingComponent from '../components/LoadingComponent';
import NotFoundPage from './error/NotFoundPage';
import SellerTransactionContainer from '../containers/SellerTransactionContainer';
import InternalServerErrorPage from './error/InternalServerErrorPage';
import BuyerTransactionContainer from '../containers/BuyerTransactionContainer';

type Props = {
  loading: boolean;
  item?: ItemData;
  load: (itemId: string) => void;
  // Logged in user info
  auth: {
    userId: number;
  };
} & RouteComponentProps<{ item_id: string }> &
  ErrorProps;

class TransactionPage extends React.Component<Props> {
  constructor(props: Props) {
    super(props);

    this.props.load(this.props.match.params.item_id);
  }

  render() {
    const {
      loading,
      item,
      auth: { userId },
    } = this.props;

    if (loading) {
      return (
        <BasePageContainer>
          <LoadingComponent />
        </BasePageContainer>
      );
    }

    if (item === undefined) {
      return <NotFoundPage />;
    }

    if (
      item.shippingStatus === undefined ||
      item.transactionEvidenceStatus === undefined ||
      item.transactionEvidenceId === undefined
    ) {
      return (
        // TODO: pass error message
        <InternalServerErrorPage />
      );
    }

    let TransactionComponent: ReactElement | undefined;

    if (userId === item.sellerId) {
      TransactionComponent = (
        <SellerTransactionContainer
          itemId={item.id}
          transactionEvidenceId={item.transactionEvidenceId}
          transactionStatus={item.transactionEvidenceStatus}
          shippingStatus={item.shippingStatus}
        />
      );
    }

    if (userId === item.buyerId) {
      TransactionComponent = (
        <BuyerTransactionContainer
          itemId={item.id}
          transactionStatus={item.transactionEvidenceStatus}
          shippingStatus={item.shippingStatus}
        />
      );
    }

    if (TransactionComponent === undefined) {
      return <NotFoundPage />;
    }

    return <BasePageContainer>{TransactionComponent}</BasePageContainer>;
  }
}

export default PageComponentWithError<Props>()(TransactionPage);
