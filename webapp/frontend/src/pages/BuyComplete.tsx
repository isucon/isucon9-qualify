import React from 'react';
import BasePageContainer from '../containers/BasePageContainer';
import { Button } from '@material-ui/core';

type Props = {
  itemId: number;
  onClickTransaction: (itemId: number) => void;
};

const BuyCompletePage: React.FC<Props> = ({ itemId, onClickTransaction }) => (
  <BasePageContainer>
    <div>購入が完了しました</div>
    <Button
      color="primary"
      variant="contained"
      onClick={(e: React.MouseEvent) => {
        onClickTransaction(itemId);
      }}
    >
      取引画面へ
    </Button>
  </BasePageContainer>
);

export default BuyCompletePage;
