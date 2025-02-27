import { FC, memo, useState } from "react";
import {
  AppBar,
  Toolbar,
  Typography,
  makeStyles,
  Avatar,
  Link,
  Button,
  IconButton,
  MenuItem,
  Menu,
} from "@material-ui/core";
import {
  PAGE_PATH_APPLICATIONS,
  PAGE_PATH_DEPLOYMENTS,
  PAGE_PATH_SETTINGS,
  PAGE_PATH_LOGIN,
  LOGOUT_ENDPOINT,
  PAGE_PATH_TOP,
  PAGE_PATH_INSIGHTS,
  PAGE_PATH_DEPLOYMENT_CHAINS,
} from "~/constants/path";
import { APP_NAME } from "~/constants/common";
import { NavLink as RouterLink } from "react-router-dom";
import ArrowDownIcon from "@material-ui/icons/ArrowDropDown";
import logo from "~~/assets/logo.svg";
import { useAppSelector } from "~/hooks/redux";

export const APP_HEADER_HEIGHT = 56;

const useStyles = makeStyles((theme) => ({
  root: {
    zIndex: theme.zIndex.drawer + 1,
    height: APP_HEADER_HEIGHT,
  },
  logo: {
    height: APP_HEADER_HEIGHT,
  },
  left: {
    flexGrow: 1,
    display: "flex",
    alignItems: "center",
  },
  appIcon: {
    marginRight: theme.spacing(2),
    width: theme.spacing(4),
    height: theme.spacing(4),
  },
  userAvatar: {
    width: theme.spacing(4),
    height: theme.spacing(4),
  },
  projectName: {
    marginLeft: theme.spacing(1),
    textTransform: "none",
  },
  right: {
    height: "100%",
    "&:hover": {
      color: theme.palette.grey[400],
    },
  },
  link: {
    marginRight: theme.spacing(2),
    display: "inline-flex",
    height: "100%",
    alignItems: "center",
    "&:hover": {
      color: theme.palette.grey[100],
      textDecoration: "none",
    },
  },
  activeLink: {
    borderBottom: `4px solid ${theme.palette.background.paper}`,
  },
}));

export const Header: FC = memo(function Header() {
  const classes = useStyles();
  const me = useAppSelector((state) => state.me);
  const [anchorEl, setAnchorEl] = useState<HTMLButtonElement | null>(null);

  const handleClose = (): void => {
    setAnchorEl(null);
  };

  return (
    <AppBar position="static" className={classes.root}>
      <Toolbar variant="dense">
        <div className={classes.left}>
          <Link
            component={RouterLink}
            to={PAGE_PATH_TOP}
            className={classes.logo}
          >
            <img className={classes.logo} src={logo} alt={APP_NAME}></img>
          </Link>
          {me?.isLogin && (
            <Button
              color="inherit"
              className={classes.projectName}
              endIcon={<ArrowDownIcon />}
            >
              {me.projectId}
            </Button>
          )}
        </div>
        <div className={classes.right}>
          {me?.isLogin ? (
            <>
              <Link
                component={RouterLink}
                className={classes.link}
                activeClassName={classes.activeLink}
                color="inherit"
                to={PAGE_PATH_APPLICATIONS}
              >
                Applications
              </Link>
              <Link
                component={RouterLink}
                className={classes.link}
                activeClassName={classes.activeLink}
                color="inherit"
                to={PAGE_PATH_DEPLOYMENTS}
              >
                Deployments
              </Link>
              <Link
                component={RouterLink}
                className={classes.link}
                activeClassName={classes.activeLink}
                color="inherit"
                to={PAGE_PATH_DEPLOYMENT_CHAINS}
              >
                Chains
              </Link>
              <Link
                component={RouterLink}
                className={classes.link}
                activeClassName={classes.activeLink}
                color="inherit"
                to={PAGE_PATH_INSIGHTS}
              >
                Insights
              </Link>
              <Link
                component={RouterLink}
                className={classes.link}
                activeClassName={classes.activeLink}
                color="inherit"
                to={PAGE_PATH_SETTINGS}
              >
                Settings
              </Link>
              <IconButton
                aria-label="User Menu"
                aria-controls="user-menu"
                aria-haspopup="true"
                onClick={(e) => setAnchorEl(e.currentTarget)}
              >
                <Avatar className={classes.userAvatar} src={me.avatarUrl} />
              </IconButton>
            </>
          ) : (
            <Link
              color="inherit"
              component={RouterLink}
              to={PAGE_PATH_LOGIN}
              className={classes.link}
            >
              <Typography variant="body2">Login</Typography>
            </Link>
          )}
        </div>
      </Toolbar>

      <Menu
        id="user-menu"
        anchorEl={anchorEl}
        open={Boolean(anchorEl)}
        onClose={handleClose}
      >
        <MenuItem component={Link} href={LOGOUT_ENDPOINT}>
          Logout
        </MenuItem>
      </Menu>
    </AppBar>
  );
});
