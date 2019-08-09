import React from 'react';
import ItemBuyFormContainer from '../containers/ItemBuyFormContainer';
import BasePageContainer from '../containers/BasePageContainer';

const ItemBuyPage: React.FC = () => {
  return (
    <BasePageContainer>
      <ItemBuyFormContainer />
    </BasePageContainer>
  );
};

export default ItemBuyPage;
