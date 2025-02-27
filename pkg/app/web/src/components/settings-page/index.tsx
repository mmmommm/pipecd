import {
  Drawer,
  List,
  ListItem,
  ListItemText,
  ListItemIcon,
  makeStyles,
  Toolbar,
  Tooltip,
} from "@material-ui/core";
import { Warning } from "@material-ui/icons";
import { FC, memo } from "react";
import { NavLink, Redirect, Route, Switch } from "react-router-dom";
import {
  PAGE_PATH_SETTINGS,
  PAGE_PATH_SETTINGS_API_KEY,
  PAGE_PATH_SETTINGS_ENV,
  PAGE_PATH_SETTINGS_PIPED,
  PAGE_PATH_SETTINGS_PROJECT,
} from "~/constants/path";
import { APIKeyPage } from "./api-key";
import { SettingsEnvironmentPage } from "./environment";
import { SettingsPipedPage } from "./piped";
import { SettingsProjectPage } from "./project";

const drawerWidth = 240;

const useStyles = makeStyles((theme) => ({
  root: {
    flex: 1,
    display: "flex",
    overflow: "hidden",
  },
  drawer: {
    width: drawerWidth,
    flexShrink: 0,
  },
  drawerPaper: {
    width: drawerWidth,
  },
  drawerContainer: {
    overflow: "auto",
  },
  content: {
    display: "flex",
    flexDirection: "column",
    flexGrow: 1,
  },
  activeNav: {
    backgroundColor: theme.palette.action.selected,
  },
  listItemIcon: {
    minWidth: 110,
  },
}));

const MENU_ITEMS = [
  ["Piped", PAGE_PATH_SETTINGS_PIPED],
  ["Environment", PAGE_PATH_SETTINGS_ENV],
  ["Project", PAGE_PATH_SETTINGS_PROJECT],
  ["API Key", PAGE_PATH_SETTINGS_API_KEY],
];

export const SettingsIndexPage: FC = memo(function SettingsIndexPage() {
  const classes = useStyles();
  return (
    <div className={classes.root}>
      <Drawer
        className={classes.drawer}
        variant="permanent"
        classes={{ paper: classes.drawerPaper }}
      >
        <Toolbar variant="dense" />
        <div className={classes.drawerContainer}>
          <List>
            {MENU_ITEMS.map(([text, link]) => (
              <ListItem
                key={`menu-item-${text}`}
                button
                component={NavLink}
                to={link}
                activeClassName={classes.activeNav}
              >
                <ListItemText primary={text} />
                {link === PAGE_PATH_SETTINGS_ENV && (
                  <ListItemIcon className={classes.listItemIcon}>
                    <Tooltip title="Deprecated. Please use Label instead.">
                      <Warning fontSize="small" />
                    </Tooltip>
                  </ListItemIcon>
                )}
              </ListItem>
            ))}
          </List>
        </div>
      </Drawer>
      <main className={classes.content}>
        <Switch>
          <Route
            exact
            path={PAGE_PATH_SETTINGS}
            component={() => <Redirect to={PAGE_PATH_SETTINGS_PIPED} />}
          />
          <Route
            exact
            path={PAGE_PATH_SETTINGS_PIPED}
            component={SettingsPipedPage}
          />
          <Route
            exact
            path={PAGE_PATH_SETTINGS_ENV}
            component={SettingsEnvironmentPage}
          />
          <Route
            exact
            path={PAGE_PATH_SETTINGS_PROJECT}
            component={SettingsProjectPage}
          />
          <Route
            exact
            path={PAGE_PATH_SETTINGS_API_KEY}
            component={APIKeyPage}
          />
        </Switch>
      </main>
    </div>
  );
});
