import React from 'react';
import ItemBuyFormContainer from '../containers/ItemBuyFormContainer';
import BasePageContainer from '../containers/BasePageContainer';
import { RouteComponentProps } from 'react-router-dom';
import { ItemData } from '../dataObjects/item';
import LoadingComponent from '../components/LoadingComponent';
import { ErrorProps, PageComponentWithError } from '../hoc/withBaseComponent';

type Props = {
  loading: boolean;
  load: (itemId: string) => void;
  item?: ItemData;
} & RouteComponentProps<{ item_id: string }> &
  ErrorProps;

class ItemBuyPage extends React.Component<Props> {
  constructor(props: Props) {
    super(props);

    const { item } = props;
    const item_id = props.match.params.item_id;

    // 商品が渡されない or 渡された商品とURLが一致しない場合は商品取得をする
    if (!item || item.id.toString() !== item_id) {
      props.load(item_id);
    }
  }

  render() {
    const { loading } = this.props;

    return (
      <BasePageContainer>
        {loading ? <LoadingComponent /> : <ItemBuyFormContainer />}
      </BasePageContainer>
    );
  }
}

export default PageComponentWithError<Props>()(ItemBuyPage);
