import React from "react";
import { ItemData } from "../dataObjects/item";
import makeStyles from "@material-ui/core/styles/makeStyles";
import { ItemListComponent } from "../components/ItemListComponent";
import SellingButtonContainer from "../containers/SellingButtonContainer";
import { ErrorProps, PageComponentWithError } from "../hoc/withBaseComponent";
import BasePageContainer from "../containers/BasePageContainer";

const useStyles = makeStyles(theme => ({
  root: {
    display: "flex",
    flexWrap: "wrap",
    marginTop: theme.spacing(1),
    justifyContent: "space-around",
    overflow: "hidden"
  }
}));

type ItemListPageProps = {
  items: ItemData[];
} & ErrorProps;

const ItemListPage: React.FC<ItemListPageProps> = ({
  items
}: ItemListPageProps) => {
  const classes = useStyles();

  return (
    <BasePageContainer>
      <div className={classes.root}>
        <ItemListComponent items={items} />
        <SellingButtonContainer />
      </div>
    </BasePageContainer>
  );
};

export default PageComponentWithError<ItemListPageProps>()(ItemListPage);
