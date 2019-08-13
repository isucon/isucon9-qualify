import React from 'react';
import { TimelineItem } from '../dataObjects/item';
import { ItemListComponent } from '../components/ItemListComponent';
import SellingButtonContainer from '../containers/SellingButtonContainer';
import { ErrorProps, PageComponentWithError } from '../hoc/withBaseComponent';
import BasePageContainer from '../containers/BasePageContainer';
import { createStyles, Theme, WithStyles } from '@material-ui/core';
import { StyleRules } from '@material-ui/core/styles';
import withStyles from '@material-ui/core/styles/withStyles';
import LoadingComponent from '../components/LoadingComponent';
import Typography from '@material-ui/core/Typography/Typography';

const styles = (theme: Theme): StyleRules =>
  createStyles({
    root: {
      display: 'flex',
      flexWrap: 'wrap',
      marginTop: theme.spacing(1),
      justifyContent: 'space-around',
      overflow: 'hidden',
    },
  });

interface ItemListPageProps extends WithStyles<typeof styles> {
  loading: boolean;
  load: () => void;
  items: TimelineItem[];
  hasNext: boolean;
  loadMore: (page: number) => void;
}

type Props = ItemListPageProps & ErrorProps;

class ItemListPage extends React.Component<Props> {
  constructor(props: Props) {
    super(props);

    this.props.load();
  }

  render() {
    const { classes, loading, items } = this.props;

    const Content: React.FC<{}> = () =>
      items.length === 0 ? (
        <div className={classes.root}>
          <Typography variant="h5">出品されている商品はありません</Typography>
        </div>
      ) : (
        <div className={classes.root}>
          <ItemListComponent {...this.props} />
          <SellingButtonContainer />
        </div>
      );

    return (
      <BasePageContainer>
        {loading ? <LoadingComponent /> : <Content />}
      </BasePageContainer>
    );
  }
}

export default PageComponentWithError<any>()(withStyles(styles)(ItemListPage));
